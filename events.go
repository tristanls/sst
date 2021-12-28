package sst

import "github.com/pkg/errors"

// NextEvent creates a singular next event.
func (s *SST) NextEvent(short, data string) (*Node, error) {
	nodes, err := s.NextEvents([]string{short}, []string{data})
	if err != nil {
		return nil, err
	}
	return nodes[0], nil
}

// NextEvents creates a set of next parallel events.
func (s *SST) NextEvents(shorts, data []string) ([]*Node, error) {
	newset := make([]*Node, 0)
	var evnt *Node
	var err error
	for i := range shorts {
		evnt, err = s.CreateNode(shorts[i], data[i], 1.0)
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

// PreviousEvents returns the previous events
func (s *SST) PreviousEvents() []*Node {
	return s.prevEvents
}
