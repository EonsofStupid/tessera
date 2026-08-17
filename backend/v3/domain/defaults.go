package domain

import (
	"github.com/EonsofStupid/tessera/backend/v3/storage/database"
	"github.com/EonsofStupid/tessera/backend/v3/storage/eventstore"
	"github.com/EonsofStupid/tessera/internal/config/systemdefaults"
	"github.com/EonsofStupid/tessera/internal/crypto"
	"github.com/EonsofStupid/tessera/internal/webauthn"
)

var (
	pool                          database.Pool
	legacyEventstore              eventstore.LegacyEventstore
	sysConfig                     systemdefaults.SystemDefaults
	passwordHasher                *crypto.Hasher
	idpEncryptionAlgo             crypto.EncryptionAlgorithm
	sessionTokenDecryptor         SessionTokenDecryptor
	mfaEncryptionAlgo             crypto.EncryptionAlgorithm
	otpSMSSecretGeneratorConfig   *crypto.GeneratorConfig
	otpEmailSecretGeneratorConfig *crypto.GeneratorConfig
	webauthnConfig                *webauthn.Config
)

func SetPool(p database.Pool) {
	pool = p
}

func SetLegacyEventstore(es eventstore.LegacyEventstore) {
	legacyEventstore = es
}

func SetSystemConfig(cfg systemdefaults.SystemDefaults) {
	sysConfig = cfg
}

func SetPasswordHasher(hasher *crypto.Hasher) {
	passwordHasher = hasher
}

func SetIDPEncryptionAlgorithm(idpEncryptionAlg crypto.EncryptionAlgorithm) {
	idpEncryptionAlgo = idpEncryptionAlg
}

func SetSessionTokenDecryptor(decryptor SessionTokenDecryptor) {
	sessionTokenDecryptor = decryptor
}

func SetOTPSMSSecretGeneratorConfig(cfg *crypto.GeneratorConfig) {
	otpSMSSecretGeneratorConfig = cfg
}

func SetWebAuthNConfig(cfg *webauthn.Config) {
	webauthnConfig = cfg
}

func SetOTPEmailSecretGeneratorConfig(cfg *crypto.GeneratorConfig) {
	otpEmailSecretGeneratorConfig = cfg
}

func SetMFAEncryptionAlgorithm(mfaEncryptionAlg crypto.EncryptionAlgorithm) {
	mfaEncryptionAlgo = mfaEncryptionAlg
}
