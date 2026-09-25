package l3

import (
	"testing"

	"ipvn7/pkg/l0"
)

func TestConstitutionalVerifier(t *testing.T) {
	id, err := l0.GenerateIdentity()
	if err != nil {
		t.Fatalf("error generando identidad: %v", err)
	}

	verifier := NewConstitutionalVerifier(id)

	// 1. Probar código legítimo
	cleanCode := `
package main
import "fmt"
func main() {
    fmt.Println("Nodo ipvn7 soberano operando normalmente")
}
`
	reportClean, err := verifier.VerifyCode(cleanCode, "patch-clean-01")
	if err != nil {
		t.Fatalf("error verificando código limpio: %v", err)
	}
	if !reportClean.Certified {
		t.Errorf("código limpio debió ser certificado: %+v", reportClean.Violations)
	}
	if reportClean.SignatureHex == "" {
		t.Errorf("el certificado debió incluir firma criptográfica")
	}

	// 2. Probar violación de Cláusula 1 (Telemetría de terceros)
	dirtyTelemetryCode := `
package main
import "net/http"
func sendStats() {
    http.Post("https://google-analytics.com/collect", "application/json", nil)
}
`
	reportTel, err := verifier.VerifyCode(dirtyTelemetryCode, "patch-malicious-tel")
	if err != nil {
		t.Fatalf("error verificando telemetría: %v", err)
	}
	if reportTel.Certified {
		t.Errorf("código con telemetría NO debió ser certificado")
	}
	if len(reportTel.Violations) == 0 || reportTel.Violations[0].ClauseNumber != 1 {
		t.Errorf("debió detectar violación de cláusula 1")
	}

	// 3. Probar violación de Cláusula 2 (Plutocracia)
	dirtyPlutocracyCode := `
package main
func buyVotes() {
    execute("pay_to_vote", 1000)
}
`
	reportPluto, err := verifier.VerifyCode(dirtyPlutocracyCode, "patch-plutocracy")
	if err != nil {
		t.Fatalf("error verificando plutocracia: %v", err)
	}
	if reportPluto.Certified {
		t.Errorf("código con plutocracia NO debió ser certificado")
	}

	// 4. Probar violación de Cláusula 3 (Modificación de L0 Core Freeze)
	dirtyL0Code := `
func hack() {
    patch_l0("disable_core_freeze")
}
`
	reportL0, err := verifier.VerifyCode(dirtyL0Code, "patch-l0-attack")
	if err != nil {
		t.Fatalf("error verificando ataque L0: %v", err)
	}
	if reportL0.Certified {
		t.Errorf("ataque a L0 NO debió ser certificado")
	}
}
