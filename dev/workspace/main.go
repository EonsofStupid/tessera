// Command workspace initializes and validates Tessera's engineering workspace.
// It never starts a service or prints a protected value.
package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	capabilitycontract "github.com/EonsofStupid/tessera/contracts/capabilities"
)

const workspaceSchema = "tessera.engineering-workspace.v1"

type manifest struct {
	SchemaVersion string       `json:"schemaVersion"`
	StateRoot     string       `json:"stateRoot"`
	BaseBranch    string       `json:"baseBranch"`
	Components    []component  `json:"components"`
	References    []reference  `json:"references"`
	Profiles      []profile    `json:"profiles"`
	Capabilities  []capability `json:"capabilities"`
	Ports         []port       `json:"ports"`
	SecretPolicy  secretPolicy `json:"secretPolicy"`
}

type reference struct {
	ID            string   `json:"id"`
	Repository    string   `json:"repository"`
	Revision      string   `json:"revision"`
	Destination   string   `json:"destination"`
	SparsePaths   []string `json:"sparsePaths"`
	RequiredPaths []string `json:"requiredPaths"`
}

type component struct {
	ID            string   `json:"id"`
	Location      string   `json:"location"`
	Authority     string   `json:"authority"`
	RequiredFiles []string `json:"requiredFiles"`
}

type profile struct {
	ID                 string   `json:"id"`
	OperatingSystems   []string `json:"operatingSystems"`
	RequiredTools      []string `json:"requiredTools"`
	RequiredComponents []string `json:"requiredComponents"`
	RequiredReferences []string `json:"requiredReferences"`
	Capabilities       []string `json:"capabilities"`
}

type capability struct {
	ID            string   `json:"id"`
	ConformanceID string   `json:"conformanceId"`
	Profiles      []string `json:"profiles"`
}

type port struct {
	Name string `json:"name"`
	Port uint16 `json:"port"`
	Bind string `json:"bind"`
}

type secretPolicy struct {
	GeneratedAtRuntime      bool     `json:"generatedAtRuntime"`
	DirectoryMode           string   `json:"directoryMode"`
	FileMode                string   `json:"fileMode"`
	ForbiddenTrackedPattern []string `json:"forbiddenTrackedPatterns"`
}

type check struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Message string `json:"message"`
	Remedy  string `json:"remedy,omitempty"`
}

type report struct {
	SchemaVersion string  `json:"schemaVersion"`
	Profile       string  `json:"profile"`
	Ready         bool    `json:"ready"`
	Checks        []check `json:"checks"`
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		printUsage(stderr)
		return 2
	}
	repositoryRoot, err := gitOutput("rev-parse", "--show-toplevel")
	if err != nil {
		fmt.Fprintf(stderr, "workspace: resolve repository root: %v\n", err)
		return 1
	}
	m, err := loadManifest(filepath.Join(repositoryRoot, "dev", "workspace", "manifest.json"))
	if err != nil {
		fmt.Fprintf(stderr, "workspace: %v\n", err)
		return 1
	}

	switch args[0] {
	case "init":
		flags := flag.NewFlagSet("init", flag.ContinueOnError)
		flags.SetOutput(stderr)
		if err := flags.Parse(args[1:]); err != nil {
			return 2
		}
		created, err := initState(repositoryRoot, m.StateRoot)
		if err != nil {
			fmt.Fprintf(stderr, "workspace init: %v\n", err)
			return 1
		}
		if created {
			fmt.Fprintf(stdout, "workspace initialized: generated protected state under %s\n", m.StateRoot)
		} else {
			fmt.Fprintf(stdout, "workspace already initialized: protected state under %s\n", m.StateRoot)
		}
		fmt.Fprintln(stdout, "no service was started and no protected value was printed")
		return 0
	case "doctor":
		flags := flag.NewFlagSet("doctor", flag.ContinueOnError)
		flags.SetOutput(stderr)
		profileID := flags.String("profile", "core", "validation profile")
		asJSON := flags.Bool("json", false, "emit machine-readable JSON")
		if err := flags.Parse(args[1:]); err != nil {
			return 2
		}
		report := doctor(repositoryRoot, m, *profileID)
		if *asJSON {
			encoder := json.NewEncoder(stdout)
			encoder.SetIndent("", "  ")
			_ = encoder.Encode(report)
		} else {
			printReport(stdout, report)
		}
		if !report.Ready {
			return 1
		}
		return 0
	case "sync-references":
		flags := flag.NewFlagSet("sync-references", flag.ContinueOnError)
		flags.SetOutput(stderr)
		referenceID := flags.String("id", "", "sync only the named reference")
		if err := flags.Parse(args[1:]); err != nil {
			return 2
		}
		if err := syncReferences(repositoryRoot, m.References, *referenceID, stdout); err != nil {
			fmt.Fprintf(stderr, "workspace sync-references: %v\n", err)
			return 1
		}
		return 0
	default:
		printUsage(stderr)
		return 2
	}
}

func printUsage(w io.Writer) {
	fmt.Fprintln(w, "usage: go run ./dev/workspace <init|doctor|sync-references> [flags]")
}

func loadManifest(path string) (manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return manifest{}, fmt.Errorf("read manifest: %w", err)
	}
	var value manifest
	if err := json.Unmarshal(data, &value); err != nil {
		return manifest{}, fmt.Errorf("decode manifest: %w", err)
	}
	if err := validateManifest(value); err != nil {
		return manifest{}, fmt.Errorf("validate manifest: %w", err)
	}
	return value, nil
}

func validateManifest(value manifest) error {
	if value.SchemaVersion != workspaceSchema {
		return fmt.Errorf("unsupported schema %q", value.SchemaVersion)
	}
	if value.StateRoot != ".artifacts/workspace" {
		return errors.New("stateRoot must remain .artifacts/workspace")
	}
	components := make(map[string]struct{}, len(value.Components))
	for _, item := range value.Components {
		if item.ID == "" || (item.Location != "self" && item.Location != "sibling") || item.Authority == "" || len(item.RequiredFiles) == 0 {
			return fmt.Errorf("component %q is incomplete", item.ID)
		}
		if _, duplicate := components[item.ID]; duplicate {
			return fmt.Errorf("duplicate component %q", item.ID)
		}
		components[item.ID] = struct{}{}
	}
	for _, required := range []string{"tessera", "vaultix", "shippin", "zuul"} {
		if _, ok := components[required]; !ok {
			return fmt.Errorf("required component %q is missing", required)
		}
	}
	references := make(map[string]struct{}, len(value.References))
	for _, item := range value.References {
		if item.ID == "" || item.Repository == "" || !isPinnedRevision(item.Revision) || len(item.SparsePaths) == 0 || len(item.RequiredPaths) == 0 {
			return fmt.Errorf("reference %q is incomplete or not pinned to a full commit", item.ID)
		}
		if !isManagedReferenceDestination(item.Destination) {
			return fmt.Errorf("reference %s destination %q must be a child of upstream", item.ID, item.Destination)
		}
		if _, duplicate := references[item.ID]; duplicate {
			return fmt.Errorf("duplicate reference %q", item.ID)
		}
		references[item.ID] = struct{}{}
		paths := append(append([]string{}, item.SparsePaths...), item.RequiredPaths...)
		for _, name := range paths {
			if !isRelativeReferencePath(name) {
				return fmt.Errorf("reference %s contains unsafe path %q", item.ID, name)
			}
		}
	}

	profiles := make(map[string]struct{}, len(value.Profiles))
	for _, item := range value.Profiles {
		if item.ID == "" || len(item.OperatingSystems) == 0 || len(item.RequiredComponents) == 0 {
			return fmt.Errorf("profile %q is incomplete", item.ID)
		}
		if _, duplicate := profiles[item.ID]; duplicate {
			return fmt.Errorf("duplicate profile %q", item.ID)
		}
		profiles[item.ID] = struct{}{}
		for _, componentID := range item.RequiredComponents {
			if _, ok := components[componentID]; !ok {
				return fmt.Errorf("profile %s references unknown component %s", item.ID, componentID)
			}
		}
		for _, referenceID := range item.RequiredReferences {
			if _, ok := references[referenceID]; !ok {
				return fmt.Errorf("profile %s references unknown upstream reference %s", item.ID, referenceID)
			}
		}
	}

	wantCapabilities := expectedCapabilities()
	gotCapabilities := make([]string, 0, len(value.Capabilities))
	seenCapabilities := make(map[string]struct{}, len(value.Capabilities))
	for _, item := range value.Capabilities {
		if item.ID == "" || item.ConformanceID == "" || len(item.Profiles) == 0 {
			return fmt.Errorf("capability %q is incomplete", item.ID)
		}
		if _, duplicate := seenCapabilities[item.ID]; duplicate {
			return fmt.Errorf("duplicate capability %q", item.ID)
		}
		seenCapabilities[item.ID] = struct{}{}
		gotCapabilities = append(gotCapabilities, item.ID)
		for _, profileID := range item.Profiles {
			if _, ok := profiles[profileID]; !ok {
				return fmt.Errorf("capability %s references unknown profile %s", item.ID, profileID)
			}
		}
	}
	sort.Strings(wantCapabilities)
	sort.Strings(gotCapabilities)
	if strings.Join(wantCapabilities, ",") != strings.Join(gotCapabilities, ",") {
		return fmt.Errorf("capability ids differ from the Tessera domain: got %v want %v", gotCapabilities, wantCapabilities)
	}
	for _, item := range value.Profiles {
		for _, capabilityID := range item.Capabilities {
			if _, ok := seenCapabilities[capabilityID]; !ok {
				return fmt.Errorf("profile %s references unknown capability %s", item.ID, capabilityID)
			}
		}
	}

	seenPorts := make(map[uint16]struct{}, len(value.Ports))
	for _, item := range value.Ports {
		if item.Name == "" || item.Port < 1024 || item.Bind != "127.0.0.1" {
			return fmt.Errorf("port allocation %q is invalid or not loopback-only", item.Name)
		}
		if _, duplicate := seenPorts[item.Port]; duplicate {
			return fmt.Errorf("duplicate port %d", item.Port)
		}
		seenPorts[item.Port] = struct{}{}
	}
	if !value.SecretPolicy.GeneratedAtRuntime || value.SecretPolicy.DirectoryMode != "0700" || value.SecretPolicy.FileMode != "0600" {
		return errors.New("secret policy must require runtime generation with 0700 directories and 0600 files")
	}
	return nil
}

func expectedCapabilities() []string {
	return capabilitycontract.Mandatory()
}

func initState(repositoryRoot, stateRoot string) (bool, error) {
	root := filepath.Join(repositoryRoot, filepath.FromSlash(stateRoot))
	if err := os.MkdirAll(root, 0o700); err != nil {
		return false, err
	}
	if err := os.Chmod(root, 0o700); err != nil {
		return false, err
	}
	for _, name := range []string{"secrets", "evidence", "logs", "state"} {
		path := filepath.Join(root, name)
		if err := os.MkdirAll(path, 0o700); err != nil {
			return false, err
		}
		if err := os.Chmod(path, 0o700); err != nil {
			return false, err
		}
	}
	masterKeyPath := filepath.Join(root, "secrets", "tessera-masterkey")
	random := make([]byte, 16)
	if _, err := rand.Read(random); err != nil {
		return false, err
	}
	file, err := os.OpenFile(masterKeyPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if errors.Is(err, os.ErrExist) {
		return false, verifySecretFile(masterKeyPath, 32)
	}
	if err != nil {
		return false, err
	}
	complete := false
	defer func() {
		_ = file.Close()
		if !complete {
			_ = os.Remove(masterKeyPath)
		}
	}()
	if _, err := io.WriteString(file, hex.EncodeToString(random)); err != nil {
		return false, err
	}
	if err := file.Sync(); err != nil {
		return false, err
	}
	if err := file.Chmod(0o600); err != nil {
		return false, err
	}
	complete = true
	return true, nil
}

func verifySecretFile(path string, expectedLength int) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if len(data) != expectedLength {
		return fmt.Errorf("protected file %s has length %d, expected %d", path, len(data), expectedLength)
	}
	if runtime.GOOS != "windows" {
		info, err := os.Stat(path)
		if err != nil {
			return err
		}
		if info.Mode().Perm() != 0o600 {
			return fmt.Errorf("protected file %s mode is %04o, expected 0600", path, info.Mode().Perm())
		}
	}
	return nil
}

func doctor(repositoryRoot string, value manifest, profileID string) report {
	result := report{SchemaVersion: workspaceSchema, Profile: profileID, Ready: true}
	selected, ok := profileByID(value.Profiles, profileID)
	if !ok {
		result.Ready = false
		result.Checks = append(result.Checks, failed("profile", "unknown profile "+profileID, "choose one of core, workspace, linux-integration, windows-ad, or release"))
		return result
	}
	if !contains(selected.OperatingSystems, runtime.GOOS) {
		result.Ready = false
		result.Checks = append(result.Checks, failed("operating-system", "profile "+profileID+" does not run on "+runtime.GOOS, "run this profile on its named self-hosted runner"))
	} else {
		result.Checks = append(result.Checks, passed("operating-system", runtime.GOOS+" is supported by "+profileID))
	}
	for _, tool := range selected.RequiredTools {
		if _, err := lookPath(tool); err != nil {
			result.Ready = false
			result.Checks = append(result.Checks, failed("tool:"+tool, tool+" is not available", "install it for the selected profile and rerun doctor"))
		} else {
			result.Checks = append(result.Checks, passed("tool:"+tool, tool+" is available"))
		}
	}

	projectsRoot, err := resolveProjectsRoot(repositoryRoot)
	if err != nil {
		result.Ready = false
		result.Checks = append(result.Checks, failed("projects-root", err.Error(), "set TESSERA_PROJECTS_ROOT to the directory containing project repositories"))
		return result
	}
	for _, componentID := range selected.RequiredComponents {
		item, found := componentByID(value.Components, componentID)
		if !found {
			result.Ready = false
			result.Checks = append(result.Checks, failed("component:"+componentID, "component is not declared", "repair dev/workspace/manifest.json"))
			continue
		}
		root := repositoryRoot
		if item.Location == "sibling" {
			root = filepath.Join(projectsRoot, item.ID)
		}
		missing := missingFiles(root, item.RequiredFiles)
		if len(missing) > 0 {
			result.Ready = false
			result.Checks = append(result.Checks, failed("component:"+item.ID, "missing contract files: "+strings.Join(missing, ", "), "restore or check out the declared sibling project"))
		} else {
			result.Checks = append(result.Checks, passed("component:"+item.ID, item.ID+" contract boundary is present"))
		}
	}
	for _, referenceID := range selected.RequiredReferences {
		item, found := referenceByID(value.References, referenceID)
		if !found {
			result.Ready = false
			result.Checks = append(result.Checks, failed("reference:"+referenceID, "reference is not declared", "repair dev/workspace/manifest.json"))
			continue
		}
		if err := verifyReference(repositoryRoot, item); err != nil {
			result.Ready = false
			result.Checks = append(result.Checks, failed("reference:"+item.ID, err.Error(), "run: go run ./dev/workspace sync-references --id "+item.ID))
		} else {
			result.Checks = append(result.Checks, passed("reference:"+item.ID, item.ID+" is present at pinned revision "+item.Revision[:12]))
		}
	}

	stateRoot := filepath.Join(repositoryRoot, filepath.FromSlash(value.StateRoot))
	if output, err := gitOutput("-C", repositoryRoot, "check-ignore", value.StateRoot); err != nil || strings.TrimSpace(output) == "" {
		result.Ready = false
		result.Checks = append(result.Checks, failed("state-ignore", value.StateRoot+" is not ignored", "add the state root to .gitignore before initializing"))
	} else {
		result.Checks = append(result.Checks, passed("state-ignore", value.StateRoot+" is ignored"))
	}
	if err := verifyProtectedDirectory(stateRoot); err != nil {
		result.Ready = false
		result.Checks = append(result.Checks, failed("state-permissions", err.Error(), "run: go run ./dev/workspace init"))
	} else if err := verifyProtectedDirectory(filepath.Join(stateRoot, "secrets")); err != nil {
		result.Ready = false
		result.Checks = append(result.Checks, failed("state-permissions", err.Error(), "run: go run ./dev/workspace init"))
	} else {
		result.Checks = append(result.Checks, passed("state-permissions", "workspace state and secrets directories have protected permissions"))
	}
	masterKeyPath := filepath.Join(stateRoot, "secrets", "tessera-masterkey")
	if err := verifySecretFile(masterKeyPath, 32); err != nil {
		result.Ready = false
		result.Checks = append(result.Checks, failed("generated-masterkey", err.Error(), "run: go run ./dev/workspace init"))
	} else {
		result.Checks = append(result.Checks, passed("generated-masterkey", "runtime-generated master key exists with protected permissions"))
	}
	if err := checkNoStaticDevelopmentMasterKey(repositoryRoot); err != nil {
		result.Ready = false
		result.Checks = append(result.Checks, failed("tracked-masterkey", err.Error(), "remove the value and consume .artifacts/workspace/secrets/tessera-masterkey"))
	} else {
		result.Checks = append(result.Checks, passed("tracked-masterkey", "development files contain no static master key"))
	}
	if err := checkNoForbiddenTrackedSecretFiles(repositoryRoot); err != nil {
		result.Ready = false
		result.Checks = append(result.Checks, failed("tracked-secret-files", err.Error(), "remove protected material from Git and generate it under the workspace state root"))
	} else {
		result.Checks = append(result.Checks, passed("tracked-secret-files", "no forbidden secret-shaped files are tracked"))
	}
	return result
}

func verifyProtectedDirectory(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("protected path %s is not a directory", path)
	}
	if runtime.GOOS != "windows" && info.Mode().Perm() != 0o700 {
		return fmt.Errorf("protected directory %s mode is %04o, expected 0700", path, info.Mode().Perm())
	}
	return nil
}

func checkNoStaticDevelopmentMasterKey(repositoryRoot string) error {
	for _, name := range []string{"dev/dev.yaml", "dev/up.sh"} {
		data, err := os.ReadFile(filepath.Join(repositoryRoot, filepath.FromSlash(name)))
		if err != nil {
			return err
		}
		text := string(data)
		if strings.Contains(text, "MasterkeyNeedsToHave32Characters") || (name == "dev/dev.yaml" && strings.Contains(text, "MasterKey:")) {
			return fmt.Errorf("%s contains a static master key", name)
		}
	}
	return nil
}

func checkNoForbiddenTrackedSecretFiles(repositoryRoot string) error {
	output, err := gitOutput("-C", repositoryRoot, "ls-files", "-z")
	if err != nil {
		return err
	}
	var forbidden []string
	for _, name := range strings.Split(output, "\x00") {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		base := filepath.Base(filepath.FromSlash(name))
		extension := strings.ToLower(filepath.Ext(base))
		secretDirectory := strings.HasPrefix(name, "secrets/") || strings.Contains(name, "/secrets/")
		secretExtension := extension == ".pem" || extension == ".key" || extension == ".p12" || extension == ".pfx"
		environmentFile := base == ".env" || strings.HasPrefix(base, ".env.")
		if secretDirectory || secretExtension || environmentFile {
			forbidden = append(forbidden, name)
		}
	}
	if len(forbidden) > 0 {
		sort.Strings(forbidden)
		return fmt.Errorf("forbidden secret-shaped paths are tracked: %s", strings.Join(forbidden, ", "))
	}
	return nil
}

func resolveProjectsRoot(repositoryRoot string) (string, error) {
	if override := strings.TrimSpace(os.Getenv("TESSERA_PROJECTS_ROOT")); override != "" {
		return filepath.Abs(override)
	}
	commonDir, err := gitOutput("-C", repositoryRoot, "rev-parse", "--git-common-dir")
	if err != nil {
		return "", fmt.Errorf("resolve Git common directory: %w", err)
	}
	if !filepath.IsAbs(commonDir) {
		commonDir = filepath.Join(repositoryRoot, commonDir)
	}
	mainRepository := filepath.Dir(filepath.Clean(commonDir))
	return filepath.Dir(mainRepository), nil
}

func gitOutput(args ...string) (string, error) {
	output, err := exec.Command("git", args...).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git %s: %s", strings.Join(args, " "), strings.TrimSpace(string(output)))
	}
	return strings.TrimSpace(string(output)), nil
}

func syncReferences(repositoryRoot string, references []reference, selectedID string, stdout io.Writer) error {
	selected := references
	if selectedID != "" {
		item, ok := referenceByID(references, selectedID)
		if !ok {
			return fmt.Errorf("unknown reference %q", selectedID)
		}
		selected = []reference{item}
	}
	for _, item := range selected {
		destination := filepath.Join(repositoryRoot, filepath.FromSlash(item.Destination))
		if info, err := os.Lstat(destination); err == nil {
			if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
				return fmt.Errorf("reference %s destination must be a real directory", item.ID)
			}
			if _, err := gitOutput("-C", destination, "rev-parse", "--git-dir"); err != nil {
				return fmt.Errorf("reference %s destination is not a Git checkout: %w", item.ID, err)
			}
			if err := verifyReferenceOrigin(destination, item); err != nil {
				return err
			}
			if dirty, err := gitOutput("-C", destination, "status", "--porcelain"); err != nil {
				return err
			} else if dirty != "" {
				return fmt.Errorf("reference %s has local changes; preserve or discard them explicitly before syncing", item.ID)
			}
		} else if !errors.Is(err, os.ErrNotExist) {
			return err
		} else if err := runGit("clone", "--filter=blob:none", "--no-checkout", item.Repository, destination); err != nil {
			return fmt.Errorf("clone reference %s: %w", item.ID, err)
		}
		if err := runGit("-C", destination, "sparse-checkout", "init", "--cone"); err != nil {
			return fmt.Errorf("initialize sparse checkout for %s: %w", item.ID, err)
		}
		args := []string{"-C", destination, "sparse-checkout", "set"}
		args = append(args, item.SparsePaths...)
		if err := runGit(args...); err != nil {
			return fmt.Errorf("configure sparse checkout for %s: %w", item.ID, err)
		}
		if err := runGit("-C", destination, "fetch", "--depth=1", "origin", item.Revision); err != nil {
			return fmt.Errorf("fetch pinned revision for %s: %w", item.ID, err)
		}
		if err := runGit("-C", destination, "checkout", "--detach", item.Revision); err != nil {
			return fmt.Errorf("check out pinned revision for %s: %w", item.ID, err)
		}
		if err := verifyReference(repositoryRoot, item); err != nil {
			return err
		}
		fmt.Fprintf(stdout, "reference synchronized: %s at %s\n", item.ID, item.Revision)
	}
	return nil
}

func verifyReference(repositoryRoot string, item reference) error {
	destination := filepath.Join(repositoryRoot, filepath.FromSlash(item.Destination))
	info, err := os.Lstat(destination)
	if err != nil {
		return fmt.Errorf("reference %s is missing", item.ID)
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
		return fmt.Errorf("reference %s destination is not a real directory", item.ID)
	}
	if err := verifyReferenceOrigin(destination, item); err != nil {
		return err
	}
	revision, err := gitOutput("-C", destination, "rev-parse", "HEAD")
	if err != nil {
		return fmt.Errorf("reference %s is not a readable Git checkout", item.ID)
	}
	if revision != item.Revision {
		return fmt.Errorf("reference %s is at %s, expected %s", item.ID, revision, item.Revision)
	}
	if dirty, err := gitOutput("-C", destination, "status", "--porcelain"); err != nil {
		return err
	} else if dirty != "" {
		return fmt.Errorf("reference %s contains local changes", item.ID)
	}
	if missing := missingFiles(destination, item.RequiredPaths); len(missing) > 0 {
		return fmt.Errorf("reference %s is missing required paths: %s", item.ID, strings.Join(missing, ", "))
	}
	return nil
}

func verifyReferenceOrigin(destination string, item reference) error {
	origin, err := gitOutput("-C", destination, "remote", "get-url", "origin")
	if err != nil {
		return fmt.Errorf("reference %s has no readable origin", item.ID)
	}
	normalize := func(value string) string {
		return strings.TrimSuffix(strings.TrimSuffix(value, "/"), ".git")
	}
	if normalize(origin) != normalize(item.Repository) {
		return fmt.Errorf("reference %s origin is %s, expected %s", item.ID, origin, item.Repository)
	}
	return nil
}

func runGit(args ...string) error {
	command := exec.Command("git", args...)
	command.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
	output, err := command.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git %s: %s", strings.Join(args, " "), strings.TrimSpace(string(output)))
	}
	return nil
}

func isPinnedRevision(value string) bool {
	if len(value) != 40 {
		return false
	}
	for _, character := range value {
		if !strings.ContainsRune("0123456789abcdef", character) {
			return false
		}
	}
	return true
}

func isManagedReferenceDestination(value string) bool {
	clean := filepath.ToSlash(filepath.Clean(filepath.FromSlash(value)))
	return clean != "upstream" && strings.HasPrefix(clean, "upstream/") && clean == value
}

func isRelativeReferencePath(value string) bool {
	clean := filepath.ToSlash(filepath.Clean(filepath.FromSlash(value)))
	return value != "" && value != "." && clean == value && !filepath.IsAbs(filepath.FromSlash(value)) && !strings.HasPrefix(clean, "../")
}

func lookPath(tool string) (string, error) {
	if runtime.GOOS == "windows" && tool == "powershell" {
		for _, candidate := range []string{"pwsh.exe", "powershell.exe"} {
			if path, err := exec.LookPath(candidate); err == nil {
				return path, nil
			}
		}
	}
	return exec.LookPath(tool)
}

func missingFiles(root string, files []string) []string {
	var missing []string
	for _, name := range files {
		if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(name))); err != nil {
			missing = append(missing, name)
		}
	}
	return missing
}

func componentByID(items []component, id string) (component, bool) {
	for _, item := range items {
		if item.ID == id {
			return item, true
		}
	}
	return component{}, false
}

func referenceByID(items []reference, id string) (reference, bool) {
	for _, item := range items {
		if item.ID == id {
			return item, true
		}
	}
	return reference{}, false
}

func profileByID(items []profile, id string) (profile, bool) {
	for _, item := range items {
		if item.ID == id {
			return item, true
		}
	}
	return profile{}, false
}

func contains(items []string, want string) bool {
	for _, item := range items {
		if item == want {
			return true
		}
	}
	return false
}

func passed(name, message string) check {
	return check{Name: name, Status: "pass", Message: message}
}

func failed(name, message, remedy string) check {
	return check{Name: name, Status: "fail", Message: message, Remedy: remedy}
}

func printReport(w io.Writer, value report) {
	for _, item := range value.Checks {
		marker := "PASS"
		if item.Status != "pass" {
			marker = "FAIL"
		}
		fmt.Fprintf(w, "%s %-24s %s\n", marker, item.Name, item.Message)
		if item.Remedy != "" {
			fmt.Fprintf(w, "     remedy: %s\n", item.Remedy)
		}
	}
	if value.Ready {
		fmt.Fprintf(w, "workspace ready for profile %s\n", value.Profile)
	} else {
		fmt.Fprintf(w, "workspace not ready for profile %s\n", value.Profile)
	}
}
