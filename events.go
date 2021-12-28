package sst

import "github.com/pkg/errors"

// NextEvent creates a singular next event.
func (s *SST) NextEvent(short, data, kind string) (*Node, error) {
	nodes, err := s.NextEvents([]string{short}, []string{data}, []string{kind})
	if err != nil {
		return nil, err
	}
	return nodes[0], nil
}

// MustNextEvent creates a singular next event, but panics on error.
func (s *SST) MustNextEvent(short, data, kind string) *Node {
	node, err := s.NextEvent(short, data, kind)
	if err != nil {
		panic(err)
	}
	return node
}

// NextEvents creates a set of next parallel events.
func (s *SST) NextEvents(shorts, data, kind []string) ([]*Node, error) {
	newset := make([]*Node, 0)
	var evnt *Node
	var err error
	for i := range shorts {
		evnt, err = s.CreateNode(shorts[i], data[i], kind[i], 1.0)
		if err != nil {
			return nil, errors.Wrapf(err, "sst: failed to create event: %v", shorts[i])
		}
		if s.prevEvents[0].Key != startEvent.Key {
			// Link all the previous events in the slice
			for j := range s.prevEvents {
				err = s.CreateLink(s.prevEvents[j], "then", evnt, 1.0)
				if err != nil {
					return nil, errors.Wrapf(err, "sst: failed to link created event: %v with %v", evnt.Key, s.prevEvents[j].Key)
				}
			}
		}
		newset = append(newset, evnt)
	}
	s.prevEvents = newset

	return newset, nil
}

// MustNextEvents creates a set of next parallel events, but panics on error.
func (s *SST) MustNextEvents(shorts, data, kind []string) []*Node {
	nodes, err := s.NextEvents(shorts, data, kind)
	if err != nil {
		panic(err)
	}
	return nodes
}

// PreviousEvents returns the previous events
func (s *SST) PreviousEvents() []*Node {
	return s.prevEvents
}
