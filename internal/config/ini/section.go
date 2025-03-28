package ini

type IniSection struct {
	name string
	keys []*IniKey // Slice to maintain order and support duplicates
}

func NewIniSection(name string) *IniSection {
	return &IniSection{
		name: name,
		keys: []*IniKey{},
	}
}

func (s *IniSection) Name() string {
	return s.name
}

func (s *IniSection) Keys() []*IniKey {
	return s.keys
}

func (s *IniSection) GetKey(name string) (string, bool) {
	for _, key := range s.keys {
		if key.Name == name {
			return key.Value, true
		}
	}
	return "", false
}

func (s *IniSection) GetKeys(name string) []string {
	var values []string
	for _, key := range s.keys {
		if key.Name == name {
			values = append(values, key.Value)
		}
	}
	return values
}

func (s *IniSection) AddKey(name, value string) {
	s.keys = append(s.keys, &IniKey{Name: name, Value: value})
	s.recalculateIndices()
}

func (s *IniSection) AddUniqueKey(name, value string) {
	for _, key := range s.keys {
		if key.Name == name && key.Value == value {
			return
		}
	}
	s.AddKey(name, value)
}

func (s *IniSection) DeleteKey(name string) {
	newKeys := []*IniKey{}
	for _, key := range s.keys {
		if key.Name != name {
			newKeys = append(newKeys, key)
		}
	}
	s.keys = newKeys
	s.recalculateIndices()
}

func (s *IniSection) DeleteUniqueKey(name string, targetValue *string, targetIndex *int) {
	newKeys := []*IniKey{}
	for i, key := range s.keys {
		if key.Name == name {
			if targetValue != nil && key.Value == *targetValue {
				continue
			}
			if targetIndex != nil && i == *targetIndex {
				continue
			}
		}
		newKeys = append(newKeys, key)
	}
	s.keys = newKeys
	s.recalculateIndices()
}

func (s *IniSection) SetKey(name, value string) {
	s.AddKey(name, value)
}

func (s *IniSection) SetUniqueKey(name, value string) {
	for _, key := range s.keys {
		if key.Name == name && key.Value == value {
			return
		}
	}

	for _, key := range s.keys {
		if key.Name == name {
			key.Value = value
			return
		}
	}
	s.AddKey(name, value)
}

func (s *IniSection) recalculateIndices() {
	for i, key := range s.keys {
		key.Index = i
	}
}
