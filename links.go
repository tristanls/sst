package sst

import (
	"context"
	"fmt"

	arango "github.com/arangodb/go-driver"
	"github.com/pkg/errors"
)

var (
	nilLink            = errors.New("sst: link is nil")
	unknownAssociation = errors.New("sst: unknown association")
)

// Link represents an edge of a Semantic Spacetime graph
type Link struct {
	// Key is a mandatory field - short name
	Key string `json:"_key"`
	// From is a mandatory field for edges
	From string `json:"_from"`
	// To is a mandatory field for edges
	To string `json:"_to"`
	// SID is a semantic ID, matches Association.Key
	SID string `json:"semantics"`
	// Negate designates a negative association, as in Not Forward or Not Backward
	Negate bool `json:"negation"`
	// Weight is the importance rank
	Weight float64 `json:"weight"`
}

// BlockLink creates the negation of the link if it does not exist or updates
// existing negated link with the new weight.
func (s *SST) BlockLink(from *Node, rel string, to *Node, weight float64) (*Link, error) {
	return s.addLink(linkFrom(from), rel, linkTo(to), weight, true)
}

// MustBlockLink invokes BlockLink, but panics on error
func (s *SST) MustBlockLink(from *Node, rel string, to *Node, weight float64) *Link {
	link, err := s.BlockLink(from, rel, to, weight)
	if err != nil {
		panic(err)
	}
	return link
}

// CreateLink creates the link if it does not exist or updates existing link
// with the new weight.
func (s *SST) CreateLink(from *Node, rel string, to *Node, weight float64) (*Link, error) {
	return s.CreateLinkByID(linkFrom(from), rel, linkTo(to), weight)
}

// MustCreateLink invokes CreateLink, but panics on error
func (s *SST) MustCreateLink(from *Node, rel string, to *Node, weight float64) *Link {
	link, err := s.CreateLink(from, rel, to, weight)
	if err != nil {
		panic(err)
	}
	return link
}

// CreateLinkByID creates the link if it does not exist or updates existing link
// with the new weight. It uses node IDs to designate link endpoints.
func (s *SST) CreateLinkByID(fromID, rel, toID string, weight float64) (*Link, error) {
	return s.addLink(fromID, rel, toID, weight, false)
}

// MustCreateLinkByID invokes CreateLinkByID, but panics on error
func (s *SST) MustCreateLinkByID(fromID, rel, toID string, weight float64) *Link {
	link, err := s.CreateLinkByID(fromID, rel, toID, weight)
	if err != nil {
		panic(err)
	}
	return link
}

// DeleteLink deletes the link if it exists.
func (s *SST) DeleteLink(from *Node, rel string, to *Node) error {
	relKey := ToDocumentKey(rel)
	association := s.associations[relKey]
	if association == nil {
		return errors.New(fmt.Sprintf("sst: invalid link type: %v", relKey))
	}
	links, err := s.linksOf(association.SemanticType)
	if err != nil {
		return err
	}
	key := linkKey(linkFrom(from), association.Key, linkTo(to))
	_, err = links.RemoveDocument(context.TODO(), key)
	if !arango.IsNotFound(err) {
		return err
	}
	return nil
}

// MustDeleteLink deletes the link if it exists, but panics on error.
func (s *SST) MustDeleteLink(from *Node, rel string, to *Node) {
	err := s.DeleteLink(from, rel, to)
	if err != nil {
		panic(err)
	}
}

// IncrementLink creates the link with weight 1.0 if it does not exist or increments
// the weight of existing link by 1.0.
func (s *SST) IncrementLink(from *Node, rel string, to *Node) (*Link, error) {
	return s.linkOp(linkFrom(from), rel, linkTo(to), 0.0, false, incrLinkOp)
}

// MustIncrementLink invokes IncrementLink, but panics on error
func (s *SST) MustIncrementLink(from *Node, rel string, to *Node) *Link {
	link, err := s.IncrementLink(from, rel, to)
	if err != nil {
		panic(err)
	}
	return link
}

// linksOf identifies links collection based on SemanticType needed
func (s *SST) linksOf(typ SemanticType) (arango.Collection, error) {
	if typ < 0 {
		typ *= -1
	}
	switch typ {
	case Near:
		return s.near, nil
	case Follows:
		return s.follows, nil
	case Contains:
		return s.contains, nil
	case Expresses:
		return s.expresses, nil
	}
	return nil, errors.New(fmt.Sprintf("sst: no link collection for semantic type: %v", int(typ)))
}

// addLink adds the link idempotently.
func (s *SST) addLink(fromID, rel, toID string, weight float64, negate bool) (*Link, error) {
	return s.linkOp(fromID, rel, toID, weight, negate, addLinkOp)
}

// addLinkOp determines link weight when adding a link. Returns weight and noop flag.
func addLinkOp(incumbent, candidate float64) (float64, bool) {
	if candidate < 0 || incumbent == candidate {
		return 0, true
	}
	return candidate, false
}

// incrLinkOp determines link weight when incrementing a link
func incrLinkOp(incumbent, candidate float64) (float64, bool) {
	return incumbent + 1.0, false
}

type linkOp func(incumbent, candidate float64) (weight float64, noop bool)

func linkFrom(n *Node) string {
	return n.Prefix + ToDocumentKey(n.Key)
}
func linkKey(from, sid, to string) string {
	return ToDocumentKey(from + sid + to)
}
func linkTo(n *Node) string {
	return linkFrom(n)
}

// linkOp creates the link or executes the designated operation on the existing link
func (s *SST) linkOp(fromID, rel, toID string, weight float64, negate bool, op linkOp) (*Link, error) {
	relKey := ToDocumentKey(rel)
	association := s.associations[relKey]
	if association == nil {
		return nil, errors.New(fmt.Sprintf("sst: invalid link type: %v", relKey))
	}
	link := &Link{
		From:   fromID,
		To:     toID,
		SID:    association.Key,
		Weight: weight,
		Negate: negate,
	}
	link.Key = linkKey(link.From, link.SID, link.To)

	links, err := s.linksOf(association.SemanticType)
	if err != nil {
		return nil, err
	}

	exists, err := links.DocumentExists(context.TODO(), link.Key)
	if err != nil {
		return nil, err
	}
	if !exists {
		_, err := links.CreateDocument(context.TODO(), link)
		if err != nil {
			return nil, errors.Wrapf(err, "sst: failed to add new link: %v", link)
		}
	} else {
		var existing Link
		_, err := links.ReadDocument(context.TODO(), link.Key, &existing)
		if err != nil {
			return nil, errors.Wrapf(err, "sst: failed to read link: %v", link.Key)
		}
		weight, noop := op(existing.Weight, link.Weight)
		if noop {
			return link, nil
		}
		link.Weight = weight
		_, err = links.UpdateDocument(context.TODO(), link.Key, link)
		if err != nil {
			return nil, errors.Wrapf(err, "sst: failed to update link: %v", link)
		}
	}
	return link, nil
}

// LinkID returns the ArangoDB _id for a link
func (s *SST) LinkID(link *Link) (string, error) {
	if link == nil {
		return "", nilLink
	}
	a := s.associations[link.SID]
	if a == nil {
		return "", unknownAssociation
	}
	return a.SemanticType.String() + "/" + a.Key, nil
}
