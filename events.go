package sst

import "github.com/pkg/errors"

// NextEvent creates a next event.
func (s *SST) NextEvent(short, data string) (*Node, error) {
	evnt, err := s.CreateNode(short, data, 1.0)
	if err != nil {
		return nil, errors.Wrapf(err, "sst: failed to create event: %v", short)
	}
	if s.prevEvent.Key != startEvent.Key {
		err = s.CreateLink(s.prevEvent, "then", evnt, 1.0)
		if err != nil {
			return evnt, errors.Wrapf(err, "sst: failed to link created event: %v with %v", short, s.prevEvent.Key)
		}
	}
	s.prevEvent = evnt
	return evnt, nil
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
		if s.prevEvent.Key != startEvent.Key {
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
	s.prevEvent = evnt

	return newset, nil
}

// PreviousEvent returns the previous event
func (s *SST) PreviousEvent() *Node {
	return s.prevEvent
}
