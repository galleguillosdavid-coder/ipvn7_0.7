package l3

import (
	"encoding/hex"
	"fmt"
	"time"
)

// seedConstitutionalAxioms establece las invariantes de la Carta Magna en ámbito "root"
func (cle *ComputationalLawEngine) seedConstitutionalAxioms() {
	axioms := []ADICORule{
		{
			ID:            "axiom:sovereignty",
			ContextDomain: "root",
			Attribute:     "all_nodes",
			Deontic:       DeonticObligation,
			Aim:           "custody_cryptographic_keys",
			Condition:     "always",
			OrElse:        "invalidation_of_did",
			Priority:      PriorityConstitutional,
			Timestamp:     time.Unix(0, 0).UTC(),
			IssuerDID:     cle.Identity.DID(),
		},
		{
			ID:            "axiom:non_aggression",
			ContextDomain: "root",
			Attribute:     "all_nodes",
			Deontic:       DeonticProhibition,
			Aim:           "inject_malicious_code_or_dos",
			Condition:     "always",
			OrElse:        "quarantine_and_slashing",
			Priority:      PriorityConstitutional,
			Timestamp:     time.Unix(0, 0).UTC(),
			IssuerDID:     cle.Identity.DID(),
		},
		{
			ID:            "axiom:semantic_neutrality",
			ContextDomain: "root",
			Attribute:     "transit_nodes",
			Deontic:       DeonticPermission,
			Aim:           "transmit_lawful_traffic",
			Condition:     "valid_pacing_and_bandwidth",
			OrElse:        "tit_for_tat_throttling",
			Priority:      PriorityConstitutional,
			Timestamp:     time.Unix(0, 0).UTC(),
			IssuerDID:     cle.Identity.DID(),
		},
		{
			ID:            "axiom:ethical_immunity",
			ContextDomain: "root",
			Attribute:     "all_nodes",
			Deontic:       DeonticProhibition,
			Aim:           "transport_destructive_malware_or_abuse",
			Condition:     "always",
			OrElse:        "permanent_did_revocation",
			Priority:      PriorityConstitutional,
			Timestamp:     time.Unix(0, 0).UTC(),
			IssuerDID:     cle.Identity.DID(),
		},
		{
			ID:            "axiom:anti_plutocracy",
			ContextDomain: "root",
			Attribute:     "senate_participants",
			Deontic:       DeonticProhibition,
			Aim:           "monetize_or_buy_voting_power",
			Condition:     "always",
			OrElse:        "annulment_of_votes_and_poc_reset",
			Priority:      PriorityConstitutional,
			Timestamp:     time.Unix(0, 0).UTC(),
			IssuerDID:     cle.Identity.DID(),
		},
		{
			ID:            "axiom:core_freeze",
			ContextDomain: "root",
			Attribute:     "all_proposals",
			Deontic:       DeonticProhibition,
			Aim:           "mutate_l0_cryptographic_primitives",
			Condition:     "always",
			OrElse:        "proposal_immediate_rejection",
			Priority:      PriorityConstitutional,
			Timestamp:     time.Unix(0, 0).UTC(),
			IssuerDID:     cle.Identity.DID(),
		},
	}

	for _, ax := range axioms {
		sig := cle.Identity.Sign([]byte(fmt.Sprintf("%s:%s:%s", ax.ID, ax.Deontic, ax.Aim)))
		ax.SignatureHex = hex.EncodeToString(sig)
		ruleCopy := ax
		cle.rules[ax.ID] = &ruleCopy

		if cle.contextRules[ax.ContextDomain] == nil {
			cle.contextRules[ax.ContextDomain] = make(map[string]*ADICORule)
		}
		cle.contextRules[ax.ContextDomain][ax.Aim] = &ruleCopy
	}
}
