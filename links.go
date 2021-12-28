package sst

import (
	"context"
	"fmt"

	arango "github.com/arangodb/go-driver"
	"github.com/pkg/errors"
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
func (s *SST) BlockLink(c1 *Node, rel string, c2 *Node, weight float64) error {
	return s.addLink(c1, rel, c2, weight, true)
}

// MustBlockLink invokes BlockLink, but panics on error
func (s *SST) MustBlockLink(c1 *Node, rel string, c2 *Node, weight float64) {
	err := s.BlockLink(c1, rel, c2, weight)
	if err != nil {
		panic(err)
	}
}

// CreateLink creates the link if it does not exist or updates existing link
// with the new weight.
func (s *SST) CreateLink(c1 *Node, rel string, c2 *Node, weight float64) error {
	return s.addLink(c1, rel, c2, weight, false)
}

// MustCreateLink invokes CreateLink, but panics on error
func (s *SST) MustCreateLink(c1 *Node, rel string, c2 *Node, weight float64) {
	err := s.CreateLink(c1, rel, c2, weight)
	if err != nil {
		panic(err)
	}
}

// IncrementLink creates the link with weight 1.0 if it does not exist or increments
// the weight of existing link by 1.0.
func (s *SST) IncrementLink(c1 *Node, rel string, c2 *Node) error {
	return s.linkOp(c1, rel, c2, 0.0, false, incrLinkOp)
}

// MustIncrementLink invokes IncrementLink, but panics on error
func (s *SST) MustIncrementLink(c1 *Node, rel string, c2 *Node) {
	err := s.IncrementLink(c1, rel, c2)
	if err != nil {
		panic(err)
	}
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
func (s *SST) addLink(c1 *Node, rel string, c2 *Node, weight float64, negate bool) error {
	return s.linkOp(c1, rel, c2, weight, negate, addLinkOp)
}

// addLinkOp determines link weight when adding a link. Returns weight and noop flag.
func addLinkOp(incumbent, candidate float64) (float64, bool) {
	if candidate < 0 || incumbent == candidate {
		return candidate, true
	}
	return candidate, false
}

// incrLinkOp determines link weight when incrementing a link
func incrLinkOp(incumbent, candidate float64) (float64, bool) {
	return incumbent + 1.0, false
}

type linkOp func(incumbent, candidate float64) (weight float64, noop bool)

// linkOp creates the link or executes the designated operation on the existing link
func (s *SST) linkOp(c1 *Node, rel string, c2 *Node, weight float64, negate bool, op linkOp) error {
	relKey := toDocumentKey(rel)
	semantics := associations[relKey]
	if semantics == nil {
		return errors.New(fmt.Sprintf("sst: invalid link type: %v", relKey))
	}
	link := &Link{
		From:   c1.Prefix + toDocumentKey(c1.Key),
		To:     c2.Prefix + toDocumentKey(c2.Key),
		SID:    semantics.Key,
		Weight: weight,
		Negate: negate,
	}
	key := toDocumentKey(link.From + link.SID + link.To)
	association := associations[link.SID]
	if association == nil {
		return errors.New(fmt.Sprintf("sst: unknown link association: %v", link.SID))
	}
	link.SID = association.Key
	link.Key = key

	links, err := s.linksOf(association.SemanticType)
	if err != nil {
		return err
	}

	exists, err := links.DocumentExists(context.TODO(), key)
	if err != nil {
		return err
	}
	if !exists {
		_, err := links.CreateDocument(context.TODO(), link)
		if err != nil {
			return errors.Wrapf(err, "sst: failed to add new link: %v", link)
		}
	} else {
		var existing Link
		_, err := links.ReadDocument(context.TODO(), key, &existing)
		if err != nil {
			return errors.Wrapf(err, "sst: failed to read link: %v", key)
		}
		weight, noop := op(existing.Weight, link.Weight)
		if noop {
			return nil
		}
		link.Weight = weight
		_, err = links.UpdateDocument(context.TODO(), key, link)
		if err != nil {
			return errors.Wrapf(err, "sst: failed to update link: %v", link)
		}
	}
	return nil
}
