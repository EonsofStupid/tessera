package domain

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

// ---- validation, blueprint-style -------------------------------------------

func validFlow() *Flow {
	return &Flow{
		Slug:        "login-password",
		Designation: DesignationAuthentication,
		Stages: []FlowStage{
			{Kind: StageIdentify},
			{Kind: StagePassword},
		},
	}
}

func TestFlowValidate_AValidFlowPasses(t *testing.T) {
	if err := validFlow().Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestFlowValidate_Refusals(t *testing.T) {
	cases := map[string]struct {
		mutate func(*Flow)
		want   []string
	}{
		"no slug": {
			func(f *Flow) { f.Slug = " " },
			[]string{"needs a slug"},
		},
		"unknown designation": {
			func(f *Flow) { f.Designation = "authentification" },
			[]string{`unknown designation "authentification"`},
		},
		"no stages": {
			func(f *Flow) { f.Stages = nil },
			[]string{"has no stages", "nothing was checked"},
		},
		"unknown stage kind": {
			func(f *Flow) { f.Stages[1].Kind = "passwrod" },
			[]string{"stage 2", `unknown kind "passwrod"`},
		},
		"identify not first": {
			func(f *Flow) { f.Stages = []FlowStage{{Kind: StagePassword}, {Kind: StageIdentify}} },
			[]string{"identify must be first"},
		},
		"never identifies": {
			func(f *Flow) { f.Stages = []FlowStage{{Kind: StagePassword}} },
			[]string{"never identifies", "against nobody"},
		},
		"identifies twice": {
			func(f *Flow) { f.Stages = append(f.Stages, FlowStage{Kind: StageIdentify}) },
			[]string{"identifies 2 times", "one subject"},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			f := validFlow()
			tc.mutate(f)
			err := f.Validate()
			if err == nil {
				t.Fatal("accepted a flow that must be refused")
			}
			for _, frag := range tc.want {
				if !strings.Contains(err.Error(), frag) {
					t.Errorf("error lacks %q:\n%v", frag, err)
				}
			}
		})
	}
}

// ---- the executor, over fake runners ---------------------------------------

// fakeRunner passes when the answer's "give" equals its kind, fails with a
// field error otherwise, and errors when told to.
type fakeRunner struct {
	kind StageKind
	fail error
}

func (r fakeRunner) Kind() StageKind { return r.kind }
func (r fakeRunner) Challenge(_ context.Context, exec *Execution, _ FlowStage) (*Challenge, error) {
	return &Challenge{
		Component: "nomen-stage-" + string(r.kind),
		Flow:      exec.Plan.FlowSlug,
		Fields:    []ChallengeField{{Name: "give", Type: "text", Required: true}},
	}, nil
}
func (r fakeRunner) Answer(_ context.Context, exec *Execution, _ FlowStage, answer map[string]any) (map[string][]string, error) {
	if r.fail != nil {
		return nil, r.fail
	}
	if answer["give"] != string(r.kind) {
		return map[string][]string{"give": {"wrong"}}, nil
	}
	if r.kind == StageIdentify {
		exec.UserID = "user-1" // what a real identify accumulates
	}
	return nil, nil
}

func testExecutor(runners ...StageRunner) *FlowExecutor {
	n := 0
	return NewFlowExecutor(time.Hour, runners...).
		WithClock(func() time.Time { return time.Unix(1000, 0) }).
		WithIDs(func() string { n++; return "id-" + string(rune('0'+n)) })
}

func TestExecutor_WalksTheWholeFlow(t *testing.T) {
	x := testExecutor(fakeRunner{kind: StageIdentify}, fakeRunner{kind: StagePassword}, fakeRunner{kind: StageTOTP})
	f := &Flow{Slug: "login-mfa", Designation: DesignationAuthentication,
		Stages: []FlowStage{{Kind: StageIdentify}, {Kind: StagePassword}, {Kind: StageTOTP}}}

	exec, ch, err := x.Start(context.Background(), "inst-1", f)
	if err != nil {
		t.Fatal(err)
	}
	if ch.Component != "nomen-stage-identify" {
		t.Fatalf("first challenge = %s", ch.Component)
	}
	if exec.ID == "" || exec.Token == "" || exec.ID == exec.Token {
		t.Fatalf("id/token = %q/%q — the token is a second secret, not the id again", exec.ID, exec.Token)
	}

	ch, err = x.Advance(context.Background(), exec, map[string]any{"give": "identify"})
	if err != nil || ch.Component != "nomen-stage-password" {
		t.Fatalf("after identify: ch=%v err=%v", ch, err)
	}
	if exec.UserID != "user-1" {
		t.Fatal("identify's accumulation did not reach the execution")
	}
	ch, err = x.Advance(context.Background(), exec, map[string]any{"give": "password"})
	if err != nil || ch.Component != "nomen-stage-totp" {
		t.Fatalf("after password: ch=%v err=%v", ch, err)
	}
	ch, err = x.Advance(context.Background(), exec, map[string]any{"give": "totp"})
	if err != nil || ch != nil {
		t.Fatalf("after last stage: ch=%v err=%v — done is nil challenge, nil error", ch, err)
	}
	if !exec.Done() {
		t.Fatal("execution not done after its last stage passed")
	}
}

func TestExecutor_AWrongAnswerStaysAndCarriesTheError(t *testing.T) {
	x := testExecutor(fakeRunner{kind: StageIdentify}, fakeRunner{kind: StagePassword})
	exec, _, err := x.Start(context.Background(), "inst-1", validFlow())
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 3; i++ { // wrong three times: same stage, every time
		ch, err := x.Advance(context.Background(), exec, map[string]any{"give": "nope"})
		if err != nil {
			t.Fatal(err)
		}
		if ch.Component != "nomen-stage-identify" || len(ch.Errors["give"]) == 0 {
			t.Fatalf("attempt %d: ch=%+v — a failed stage re-asks with its errors", i, ch)
		}
		if exec.Position != 0 {
			t.Fatalf("attempt %d moved the position to %d", i, exec.Position)
		}
	}
	// And the right answer still works after the wrong ones.
	if _, err := x.Advance(context.Background(), exec, map[string]any{"give": "identify"}); err != nil {
		t.Fatal(err)
	}
	if exec.Position != 1 {
		t.Fatal("the right answer did not advance")
	}
}

// Infrastructure failure and wrong answer are different things: one is the
// client's to fix, the other is ours, and conflating them tells the customer
// their password is wrong when the database is down.
func TestExecutor_InfrastructureFailureIsNotAFieldError(t *testing.T) {
	boom := errors.New("connection refused")
	x := testExecutor(fakeRunner{kind: StageIdentify, fail: boom}, fakeRunner{kind: StagePassword})
	exec, _, err := x.Start(context.Background(), "inst-1", validFlow())
	if err != nil {
		t.Fatal(err)
	}
	_, err = x.Advance(context.Background(), exec, map[string]any{"give": "identify"})
	if !errors.Is(err, boom) {
		t.Fatalf("err = %v, want the underlying failure to surface", err)
	}
	if exec.Position != 0 {
		t.Fatal("a failed lookup must not advance anything")
	}
}

func TestExecutor_ExpiryAndCompletionAreTypedRefusals(t *testing.T) {
	x := testExecutor(fakeRunner{kind: StageIdentify}, fakeRunner{kind: StagePassword})
	exec, _, err := x.Start(context.Background(), "inst-1", validFlow())
	if err != nil {
		t.Fatal(err)
	}

	// Expired: the refusal points at start, because that is the fix.
	late := func() time.Time { return time.Unix(1000, 0).Add(2 * time.Hour) }
	if _, err := x.WithClock(late).Advance(context.Background(), exec, nil); !errors.Is(err, ErrExecutionExpired) {
		t.Fatalf("err = %v, want ErrExecutionExpired", err)
	}

	// Completed: answering again is a distinct refusal, not a re-run.
	x = testExecutor(fakeRunner{kind: StageIdentify}, fakeRunner{kind: StagePassword})
	exec, _, _ = x.Start(context.Background(), "inst-1", validFlow())
	_, _ = x.Advance(context.Background(), exec, map[string]any{"give": "identify"})
	_, _ = x.Advance(context.Background(), exec, map[string]any{"give": "password"})
	if _, err := x.Advance(context.Background(), exec, map[string]any{"give": "password"}); !errors.Is(err, ErrExecutionDone) {
		t.Fatalf("err = %v, want ErrExecutionDone", err)
	}
}

// A flow that would strand a customer at stage three is refused at stage zero.
func TestExecutor_StartRefusesAPlanNobodyCanRun(t *testing.T) {
	x := testExecutor(fakeRunner{kind: StageIdentify}) // no password runner
	_, _, err := x.Start(context.Background(), "inst-1", validFlow())
	if !errors.Is(err, ErrNoRunner) {
		t.Fatalf("err = %v, want ErrNoRunner before any challenge is issued", err)
	}
}

func TestNewFlowExecutor_DuplicateKindPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("two runners for one kind must panic at construction")
		}
	}()
	NewFlowExecutor(time.Hour, fakeRunner{kind: StagePassword}, fakeRunner{kind: StagePassword})
}
