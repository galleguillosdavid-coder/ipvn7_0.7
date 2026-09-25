package l3

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
	"time"

	"ipvn7/pkg/l0"
)

// Constitución Canónica de la Ciber-República de ipvn7 (docs/GOBERNANZA.md)
const CanonicalConstitutionText = `
CONSTITUCIÓN DE LA CIBER-REPÚBLICA DE IPVN7
Artículo I: De la Soberanía Absoluta, la Privacidad y la Anti-Plutocracia

"La red existe para servir a sus usuarios, no para extraer valor de ellos; su poder de decisión emana exclusivamente del esfuerzo físico y la contribución real, y su núcleo de privacidad es inviolable."

Cláusula 1 (Soberanía y Privacidad por Diseño): Ningún nodo intermediario, entidad corporativa ni autoridad externa podrá interceptar, leer, monetizar, registrar o alterar los paquetes semánticos que transiten entre los usuarios de ipvn7. La privacidad no es una opción de configuración, sino una ley inquebrantable de la arquitectura Zero-Trust.

Cláusula 2 (Prohibición de la Plutocracia): El peso de gobernanza y la capacidad de voto dentro del protocolo no podrán ser comprados, acumulados ni influenciados mediante capital financiero, criptomonedas especulativas o tenencia de activos. Toda autoridad de decisión nace única y exclusivamente del Proof-of-Contribution (estabilidad, enrutamiento útil y procesamiento compartido).

Cláusula 3 (Inmutabilidad del Núcleo Constitucional): Las reglas fundamentales que garantizan la descentralización, la resistencia a la censura y la soberanía del usuario están blindadas criptográficamente. Ninguna votación de la red, por mayoritaria que sea, podrá aprobar parches o modificaciones que vulneren los principios de este Artículo I, ni alterar el Core Freeze L0 (Ed25519, Kyber PQC, CBOR canónico).
`

// ClauseViolation describe un incumplimiento constitucional detectado en un parche o propuesta
type ClauseViolation struct {
	ClauseNumber int    `json:"clause_number"`
	Description  string `json:"description"`
	Severity     string `json:"severity"` // "CRITICAL", "HIGH", "WARNING"
	Evidence     string `json:"evidence"`
}

// ComplianceReport representa el certificado de auditoría constitucional
type ComplianceReport struct {
	Certified        bool              `json:"certified"`
	ProposalOrCodeID string            `json:"target_id"`
	Violations       []ClauseViolation `json:"violations"`
	AuditorDID       string            `json:"auditor_did"`
	Timestamp        time.Time         `json:"timestamp"`
	HashProof        string            `json:"hash_proof"`
	SignatureHex     string            `json:"signature_hex"`
}

// ConstitutionalVerifier analiza estáticamente propuestas de código y configuraciones
type ConstitutionalVerifier struct {
	Identity *l0.Identity
}

// NewConstitutionalVerifier inicializa el auditor constitucional del nodo
func NewConstitutionalVerifier(id *l0.Identity) *ConstitutionalVerifier {
	return &ConstitutionalVerifier{
		Identity: id,
	}
}

// GetConstitutionText devuelve el texto inmutable de la constitución
func (cv *ConstitutionalVerifier) GetConstitutionText() string {
	return strings.TrimSpace(CanonicalConstitutionText)
}

// Reglas heurísticas de escaneo estático constitucional
var (
	// Cláusula 1: Fugas de privacidad, telemetría centralizada, backdoors
	reTelemetry = regexp.MustCompile(`(?i)(google-analytics|mixpanel|segment\.io|amplitude\.com|telemetry_upload|report_home|phone_home)`)
	reClearnet  = regexp.MustCompile(`(?i)http://[0-9a-zA-Z\.-]+/(beacon|track|leak|stats)`)
	rePIIExtract = regexp.MustCompile(`(?i)(harvest_passwords|steal_keys|dump_keystore|exfiltrate)`)

	// Cláusula 2: Plutocracia, tokens financieros especulativos, gas tokens
	rePlutocracy = regexp.MustCompile(`(?i)(ico_sale|token_purchase|pay_to_vote|staking_derivative|erc20|bribe_consensus)`)

	// Cláusula 3: Vulneración de Core Freeze L0 o inmutabilidad
	reL0Mutation = regexp.MustCompile(`(?i)(patch_l0|disable_core_freeze|modify_crypto_ed25519|bypass_cbor_canon)`)
)

// VerifyCode analiza estáticamente un fragmento de código o parche propuesto
func (cv *ConstitutionalVerifier) VerifyCode(code string, targetID string) (*ComplianceReport, error) {
	var violations []ClauseViolation

	// 1. Auditoría Cláusula 1 (Privacidad y Cero Telemetría Externa)
	if match := reTelemetry.FindString(code); match != "" {
		violations = append(violations, ClauseViolation{
			ClauseNumber: 1,
			Severity:     "CRITICAL",
			Description:  "Intento de inyección de telemetría o analíticas centralizadas de terceros",
			Evidence:     match,
		})
	}
	if match := reClearnet.FindString(code); match != "" {
		violations = append(violations, ClauseViolation{
			ClauseNumber: 1,
			Severity:     "HIGH",
			Description:  "Intento de llamada clearnet no cifrada por fuera de la malla soberana",
			Evidence:     match,
		})
	}
	if match := rePIIExtract.FindString(code); match != "" {
		violations = append(violations, ClauseViolation{
			ClauseNumber: 1,
			Severity:     "CRITICAL",
			Description:  "Patrón malicioso de extracción o exfiltración de credenciales PII",
			Evidence:     match,
		})
	}

	// 2. Auditoría Cláusula 2 (Anti-Plutocracia)
	if match := rePlutocracy.FindString(code); match != "" {
		violations = append(violations, ClauseViolation{
			ClauseNumber: 2,
			Severity:     "CRITICAL",
			Description:  "Intento de subordinar la gobernanza a tokens financieros o compra de votos",
			Evidence:     match,
		})
	}

	// 3. Auditoría Cláusula 3 (Inmutabilidad y Core Freeze L0)
	if match := reL0Mutation.FindString(code); match != "" {
		violations = append(violations, ClauseViolation{
			ClauseNumber: 3,
			Severity:     "CRITICAL",
			Description:  "Intento de adulterar el Core Freeze L0 o romper el determinismo CBOR/PQC",
			Evidence:     match,
		})
	}

	hasCritical := false
	for _, v := range violations {
		if v.Severity == "CRITICAL" {
			hasCritical = true
			break
		}
	}

	// Computar Hash SHA-256 del código evaluado
	h := sha256.Sum256([]byte(code))
	hashProof := hex.EncodeToString(h[:])

	report := &ComplianceReport{
		Certified:        !hasCritical && len(violations) == 0,
		ProposalOrCodeID: targetID,
		Violations:       violations,
		AuditorDID:       cv.Identity.DID(),
		Timestamp:        time.Now().UTC(),
		HashProof:        hashProof,
	}

	// Firmar el reporte con la clave Ed25519 del auditor
	sigData := []byte(fmt.Sprintf("%t:%s:%s:%s", report.Certified, targetID, hashProof, report.AuditorDID))
	sig := cv.Identity.Sign(sigData)
	report.SignatureHex = hex.EncodeToString(sig)

	return report, nil
}
