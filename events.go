package sst

import "github.com/pkg/errors"

// NextEvent creates a singular next event.
func (s *SST) NextEvent(kind, key string, data map[string]interface{}) (*Node, error) {
	nodes, err := s.NextEvents([]string{kind}, []string{key}, []map[string]interface{}{data})
	if err != nil {
		return nil, err
	}
	return nodes[0], nil
}

// MustNextEvent creates a singular next event, but panics on error.
func (s *SST) MustNextEvent(kind, key string, data map[string]interface{}) *Node {
	node, err := s.NextEvent(kind, key, data)
	if err != nil {
		panic(err)
	}
	return node
}

// NextEvents creates a set of next parallel events.
func (s *SST) NextEvents(kind, keys []string, data []map[string]interface{}) ([]*Node, error) {
	newset := make([]*Node, 0)
	var evnt *Node
	var err error
	for i := range keys {
		evnt, err = s.CreateNode(kind[i], keys[i], data[i], 1.0)
		if err != nil {
			return nil, errors.Wrapf(err, "sst: failed to create event: %v", keys[i])
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
func (s *SST) MustNextEvents(kind, keys []string, data []map[string]interface{}) []*Node {
	nodes, err := s.NextEvents(kind, keys, data)
	if err != nil {
		panic(err)
	}
	return nodes
}

// PreviousEvents returns the previous events
func (s *SST) PreviousEvents() []*Node {
	return s.prevEvents
}
