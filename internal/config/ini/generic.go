package ini

import (
	"bufio"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/spf13/cast"
)

type GenericIniFile struct {
	Sections   []*IniSection  // Ordered list of sections
	SectionMap map[string]int // Map of section name to its index in Sections slice
}

func NewGenericIniFile() *GenericIniFile {
	return &GenericIniFile{
		Sections:   []*IniSection{},
		SectionMap: make(map[string]int),
	}
}

func (f *GenericIniFile) GetSection(name string) *IniSection {
	if idx, exists := f.SectionMap[name]; exists {
		return f.Sections[idx]
	}
	return nil
}

func (f *GenericIniFile) AddSection(name string) *IniSection {
	if idx, exists := f.SectionMap[name]; exists {
		return f.Sections[idx]
	}

	section := &IniSection{
		Name: name,
		Keys: []*IniKey{},
	}
	f.Sections = append(f.Sections, section)
	f.SectionMap[name] = len(f.Sections) - 1
	return section
}

func (f *GenericIniFile) DeleteSection(name string) bool {
	if idx, exists := f.SectionMap[name]; exists {
		f.Sections = append(f.Sections[:idx], f.Sections[idx+1:]...)
		delete(f.SectionMap, name)

		for i := idx; i < len(f.Sections); i++ {
			f.SectionMap[f.Sections[i].Name] = i
		}
	}
	return f.GetSection(name) == nil
}

func (f *GenericIniFile) GetKey(section string, key string, defvalue string) string {
	sect := f.GetSection(section)
	if sect != nil {
		if value, exists := sect.GetKey(key); exists {
			return value
		}
	}
	return defvalue
}

func (f *GenericIniFile) GetKeyBool(section string, key string, defvalue bool) bool {
	defvalueStr := cast.ToString(defvalue)
	value := f.GetKey(section, key, defvalueStr)
	if value != defvalueStr {
		return cast.ToBool(value)
	}
	return defvalue
}

func (f *GenericIniFile) GetKeyInt(section string, key string, defvalue int) int {
	defvalueStr := cast.ToString(defvalue)
	value := f.GetKey(section, key, defvalueStr)
	if value != defvalueStr {
		return cast.ToInt(value)
	}
	return defvalue
}

func (f *GenericIniFile) GetKeyFloat(section string, key string, defvalue float64) float64 {
	defvalueStr := cast.ToString(defvalue)
	value := f.GetKey(section, key, defvalueStr)
	if value != defvalueStr {
		return cast.ToFloat64(value)
	}
	return defvalue
}

func (f *GenericIniFile) GetKeys(section string, key string) []string {
	sect := f.GetSection(section)
	if sect != nil {
		return sect.GetKeys(key)
	}
	return nil
}

func (f *GenericIniFile) SetKey(section string, key string, value string, unique bool) bool {
	return f.setInterface(section, key, value, unique)
}

func (f *GenericIniFile) SetKeyBool(section string, key string, value bool, unique bool) bool {
	return f.setInterface(section, key, value, unique)
}

func (f *GenericIniFile) SetKeyInt(section string, key string, value int, unique bool) bool {
	return f.setInterface(section, key, value, unique)
}

func (f *GenericIniFile) SetKeyFloat(section string, key string, value float64, unique bool) bool {
	return f.setInterface(section, key, value, unique)
}

func (f *GenericIniFile) DeleteKey(section string, key string) bool {
	sect := f.GetSection(section)
	if sect != nil {
		sect.DeleteKey(key)
		_, exists := sect.GetKey(key)
		return !exists
	}
	return false
}

func (f *GenericIniFile) DeleteUniqueKey(section string, key string, targetValue *string, targetIndex *int) bool {
	sect := f.GetSection(section)
	if sect != nil {
		keyCountBefore := len(sect.GetKeys(key))
		sect.DeleteUniqueKey(key, targetValue, targetIndex)
		keyCountAfter := len(sect.GetKeys(key))
		return keyCountAfter == (keyCountBefore - 1)
	}
	return false
}

func (f *GenericIniFile) Load(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file '%s': %v", filePath, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var currentSection *IniSection

	for scanner.Scan() {
		line := scanner.Text()

		// Ignore comments and empty lines
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, ";") || strings.HasPrefix(trimmed, "#") {
			continue
		}

		// Check for section header
		if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") {
			sectionName := strings.TrimSpace(trimmed[1 : len(trimmed)-1])
			currentSection = f.AddSection(sectionName)
		} else if currentSection != nil {
			// Parse key/value pair
			parts := strings.SplitN(trimmed, "=", 2)
			if len(parts) != 2 {
				return fmt.Errorf("invalid line in file '%s': %s", filePath, line)
			}
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			currentSection.AddKey(key, value)
		} else {
			return fmt.Errorf("key-value pair found outside of a section in file '%s': %s", filePath, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading file '%s': %v", filePath, err)
	}
	return nil
}

func (f *GenericIniFile) Save(filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file '%s': %v", filePath, err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, section := range f.Sections {
		// Write section header
		if section.Name != "" {
			_, _ = writer.WriteString(fmt.Sprintf("[%s]\n", section.Name))
		}

		// Write each key
		for _, key := range section.Keys {
			_, _ = writer.WriteString(fmt.Sprintf("%s=%s\n", key.Name, key.Value))
		}

		// Add a newline between sections
		_, _ = writer.WriteString("\n")
	}
	writer.Flush()
	return nil
}

func (f *GenericIniFile) setInterface(section string, key string, value interface{}, unique bool) bool {
	sect := f.GetSection(section)

	// Add the section if it doesn't exists
	if sect == nil {
		sect = f.AddSection(section)
	}

	var valueStr = cast.ToString(value)
	var isSet bool

	if unique {
		sect.SetUniqueKey(key, valueStr)
		_, isSet = sect.GetKey(key)
	} else {
		sect.SetKey(key, valueStr)
		allKeys := sect.GetKeys(key)
		isSet = slices.Contains(allKeys, valueStr)
	}
	return isSet
}
