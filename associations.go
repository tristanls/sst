package sst

import (
	"fmt"

	"github.com/pkg/errors"
)

type SemanticType int

const (
	Near SemanticType = iota
	Follows
	Contains
	Expresses
)

func (a SemanticType) String() string {
	switch a {
	case Near:
		return "Near"
	case Follows:
		return "Follows"
	case -Follows:
		return "Precedes"
	case Contains:
		return "Contains"
	case -Contains:
		return "Constitutes"
	case Expresses:
		return "Expresses"
	case -Expresses:
		return "Describes"
	}
	return "unknown"
}

// Association stores invariant relationship data as lookup tables to
// reduce database storage.
type Association struct {
	Key          string       `json:"_key"`
	SemanticType SemanticType `json:"stype"`
	Fwd          string       `json:"fwd"`
	Bwd          string       `json:"bwd"`
	Nfwd         string       `json:"nfwd"`
	Nbwd         string       `json:"nbwd"`
}

var (
	// default associations
	associations = map[string]*Association{
		"contains":        {"contains", Contains, "contains", "belongs to or is part of", "does not contain", "is not part of"},
		"generalizes":     {"generalizes", Contains, "generalizes", "is a special case of", "is not a generalization of", "is not a special case of"},
		"part_of":         {"part_of", -Contains, "is part of", "incorporates", "is not part of", "doesn't incorporate"},
		"has_role":        {"has_role", Expresses, "has the role of", "is a role fulfilled by", "has no role", "is not a role fulfilled by"},
		"originates_from": {"originates_from", Follows, "originates from", "is the source/origin of", "does not originate from", "is not the source/origin of"},
		"expresses":       {"expresses", Expresses, "expresses an attribute", "is an attribute of", "has no attribute", "is not an attribute of"},
		"promises":        {"promises", Expresses, "promises/intends", "is intended/promised by", "rejects/promises to not", "is rejected by"},
		"has_name":        {"has_name", Expresses, "has proper name", "is the proper name of", "is not named", "isn't the proper name of"},
		"follows_from":    {"follows_from", Follows, "follows on from", "is followed by", "does not follow", "does not precede"},
		"uses":            {"uses", Follows, "uses", "is used by", "does not use", "is not used by"},
		"caused_by":       {"caused_by", Follows, "caused by", "may cause", "was not caused by", "probably didn't cause"},
		"derives_from":    {"derives_from", Follows, "derives from", "leads to", "does not derive from", "does not lead to"},
		"depends":         {"depends", Follows, "may depends on", "may determine", "doesn't depend on", "doesn't determine"},
		"next":            {"next", -Follows, "comes before", "comes after", "is not before", "is not after"},
		"then":            {"then", -Follows, "then", "previously", "but not", "didn't follow"},
		"leads_to":        {"leads_to", -Follows, "leads to", "doesn't imply", "doesn't reach", "doesn't precede"},
		"precedes":        {"precedes", -Follows, "precedes", "follows", "doesn't precede", "doesn't follow"},
		"related":         {"related", Near, "may be related to", "may be related to", "likely unrelated to", "likely unrelated to"},
		"alias":           {"alias", Near, "also known as", "also known as", "not known as", "not known as"},
		"is_like":         {"is_like", Near, "is similar to", "is similar to", "is unlike", "is unlike"},
		"connected":       {"connected", Near, "is connected to", "is connected to", "is not connected to", "is not connected to"},
		"coactive":        {"coactive", Near, "occurred together with", "occurred together with", "never appears with", "never appears with"},
	}
)

// CreateAssociation creates a new association
func (s *SST) CreateAssociation(a *Association) error {
	a.Key = ToDocumentKey(a.Key)
	existing := s.associations[a.Key]
	if existing == nil {
		s.associations[a.Key] = a
		return nil
	}
	if existing == a {
		return nil
	}
	return errors.New(fmt.Sprintf("sst: failed to create association %v due to existing association %v", a, existing))
}

// MustCreateAssociation creates a new association, panics on error
func (s *SST) MustCreateAssociation(a *Association) {
	err := s.CreateAssociation(a)
	if err != nil {
		panic(err)
	}
}
