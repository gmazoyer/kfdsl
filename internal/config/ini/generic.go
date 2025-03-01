package ini

import (
	"bufio"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/K4rian/dslogger"
	"github.com/spf13/cast"

	"github.com/K4rian/kfdsl/internal/log"
)

type GenericIniFile struct {
	name            string
	sections        []*IniSection   // Ordered list of sections
	sectionMap      map[string]int  // Map of section name to its index in Sections slice
	normalizedNames map[string]bool // Tracks lowercase section names to prevent duplicates
	Logger          *dslogger.Logger
}

func NewGenericIniFile(name string) *GenericIniFile {
	return &GenericIniFile{
		name:            name,
		sections:        []*IniSection{},
		sectionMap:      make(map[string]int),
		normalizedNames: make(map[string]bool),
		Logger:          log.Logger.WithService(name),
	}
}

func (f *GenericIniFile) Name() string {
	return f.name
}

func (f *GenericIniFile) Sections() []*IniSection {
	return f.sections
}

func (f *GenericIniFile) GetSection(name string) *IniSection {
	if _, exists := f.normalizedNames[strings.ToLower(name)]; !exists {
		return nil
	}

	if idx, exists := f.sectionMap[name]; exists {
		return f.sections[idx]
	}
	return nil
}

func (f *GenericIniFile) AddSection(name string) (*IniSection, error) {
	lowerName := strings.ToLower(name)
	if _, exists := f.normalizedNames[strings.ToLower(name)]; exists {
		return nil, fmt.Errorf("duplicate section found: %s", name)
	}

	section := NewIniSection(name)
	f.sections = append(f.sections, section)
	f.sectionMap[name] = len(f.sections) - 1
	f.normalizedNames[lowerName] = true

	f.Logger.Debug("Adding new section",
		"function", "AddSection", "section", name, "totalSections", len(f.sections))
	return section, nil
}

func (f *GenericIniFile) DeleteSection(name string) bool {
	lowerName := strings.ToLower(name)
	idx, exists := f.sectionMap[lowerName]
	if !exists {
		return false
	}

	f.sections = slices.Delete(f.sections, idx, idx+1)
	delete(f.sectionMap, lowerName)
	delete(f.normalizedNames, lowerName)

	// Rebuild the map
	for i, section := range f.sections {
		f.sectionMap[strings.ToLower(section.Name())] = i
	}

	f.Logger.Debug("Deleting section",
		"function", "DeleteSection", "section", name, "totalSections", len(f.sections))
	return true
}

func (f *GenericIniFile) GetKey(section string, key string, defvalue string) string {
	if sect := f.GetSection(section); sect != nil {
		if value, exists := sect.GetKey(key); exists {
			return value
		}
	}
	return defvalue
}

func (f *GenericIniFile) GetKeyBool(section string, key string, defvalue bool) bool {
	value := f.GetKey(section, key, "")
	if result, err := cast.ToBoolE(value); err == nil {
		return result
	}
	return defvalue
}

func (f *GenericIniFile) GetKeyInt(section string, key string, defvalue int) int {
	value := f.GetKey(section, key, "")
	if result, err := cast.ToIntE(value); err == nil {
		return result
	}
	return defvalue
}

func (f *GenericIniFile) GetKeyFloat(section string, key string, defvalue float64) float64 {
	value := f.GetKey(section, key, "")
	if result, err := cast.ToFloat64E(value); err == nil {
		return result
	}
	return defvalue
}

func (f *GenericIniFile) GetKeys(section string, key string) []string {
	if sect := f.GetSection(section); sect != nil {
		return sect.GetKeys(key)
	}
	return nil
}

func (f *GenericIniFile) HasKey(section string, key string) bool {
	if sect := f.GetSection(section); sect != nil {
		_, exists := sect.GetKey(key)
		return exists
	}
	return false
}

func (f *GenericIniFile) SetKey(section string, key string, value string, unique bool) bool {
	return f.setKeyValue(section, key, value, unique)
}

func (f *GenericIniFile) SetKeyBool(section string, key string, value bool, unique bool) bool {
	return f.setKeyValue(section, key, value, unique)
}

func (f *GenericIniFile) SetKeyInt(section string, key string, value int, unique bool) bool {
	return f.setKeyValue(section, key, value, unique)
}

func (f *GenericIniFile) SetKeyFloat(section string, key string, value float64, unique bool) bool {
	return f.setKeyValue(section, key, value, unique)
}

func (f *GenericIniFile) DeleteKey(section string, key string) bool {
	if !f.HasKey(section, key) {
		return false
	}

	if sect := f.GetSection(section); sect != nil {
		val, _ := sect.GetKey(key)

		sect.DeleteKey(key)

		// The key shouldn't exists anymore
		_, exists := sect.GetKey(key)
		if !exists {
			f.Logger.Debug("Deleting key",
				"function", "DeleteKey", "section", section, "key", key, "value", val)
			return true
		}
	}
	return false
}

func (f *GenericIniFile) DeleteUniqueKey(section string, key string, targetValue *string, targetIndex *int) bool {
	if !f.HasKey(section, key) {
		return false
	}

	if sect := f.GetSection(section); sect != nil {
		keyCountBefore := len(sect.GetKeys(key))
		sect.DeleteUniqueKey(key, targetValue, targetIndex)
		keyCountAfter := len(sect.GetKeys(key))

		if keyCountAfter == (keyCountBefore - 1) {
			fields := []any{
				"function", "DeleteUniqueKey",
				"section", section,
				"key", key,
			}
			if targetValue != nil {
				fields = append(fields, "value", *targetValue)
			}
			if targetIndex != nil {
				fields = append(fields, "index", *targetIndex)
			}
			f.Logger.Debug("Deleting unique key",
				fields...)
			return true
		}
	}
	return false
}

func (f *GenericIniFile) Load(filePath string) error {
	f.Logger.Debug("Loading ini file",
		"function", "Load", "file", filePath)

	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file '%s': %v", filePath, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var currentSection *IniSection

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, ";") || strings.HasPrefix(line, "#") {
			continue
		}

		// Check for section header
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			sectionName := strings.TrimSpace(line[1 : len(line)-1])
			if currentSection, err = f.AddSection(sectionName); err != nil {
				return err
			}
		} else if currentSection != nil {
			// Parse key/value pair
			parts := strings.SplitN(line, "=", 2)
			if len(parts) != 2 {
				return fmt.Errorf("invalid line in file '%s': %s", filePath, line)
			}

			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])
			currentSection.AddKey(key, val)

			f.Logger.Debug("Parsing key",
				"function", "Load", "section", currentSection.Name(), "key", key, "value", val)
		} else {
			return fmt.Errorf("key-value pair found outside of a section in file '%s': %s", filePath, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading file '%s': %v", filePath, err)
	}

	f.Logger.Debug("Ini file successfully loaded",
		"function", "Save", "file", filePath)
	return nil
}

func (f *GenericIniFile) Save(filePath string) error {
	tempFilePath := filePath + ".tmp"

	f.Logger.Debug("Creating temp ini file",
		"function", "Save", "file", tempFilePath)

	file, err := os.Create(tempFilePath)
	if err != nil {
		return fmt.Errorf("failed to create file '%s': %v", tempFilePath, err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			f.Logger.Error("failed to close file",
				"function", "Save", "file", tempFilePath, "error", closeErr)
		}
	}()

	writer := bufio.NewWriter(file)
	for _, section := range f.sections {
		// Write section header
		if section.Name() != "" {
			f.Logger.Debug("Writing section",
				"function", "Save", "section", section.Name())

			if _, err := writer.WriteString(fmt.Sprintf("[%s]\n", section.Name())); err != nil {
				return fmt.Errorf("failed to write section header for '%s': %v", section.Name(), err)
			}
		}

		// Write each key
		for _, key := range section.Keys() {
			f.Logger.Debug("Writing key",
				"function", "Save", "section", section.Name(), "key", key.Name, "value", key.Value)

			if _, err := writer.WriteString(fmt.Sprintf("%s=%s\n", key.Name, key.Value)); err != nil {
				return fmt.Errorf("failed to write key '%s' in section '%s': %v", key.Name, section.Name(), err)
			}
		}

		// Add a newline between sections
		if _, err := writer.WriteString("\n"); err != nil {
			return fmt.Errorf("failed to write newline between sections: %v", err)
		}
	}

	if err = writer.Flush(); err != nil {
		return fmt.Errorf("failed to flush file '%s': %v", tempFilePath, err)
	}

	if err = file.Sync(); err != nil {
		return fmt.Errorf("failed to sync file '%s': %v", tempFilePath, err)
	}

	f.Logger.Debug("Temp ini file successfully saved",
		"function", "Save", "file", tempFilePath)

	if err := os.Rename(tempFilePath, filePath); err != nil {
		return fmt.Errorf("failed to rename file '%s' to '%s': %v", tempFilePath, filePath, err)
	}

	f.Logger.Debug("Ini file successfully saved",
		"function", "Save", "sourcefile", tempFilePath, "destFile", filePath)
	return nil
}

func (f *GenericIniFile) setKeyValue(section string, key string, value any, unique bool) bool {
	sect := f.GetSection(section)

	// Add the section if it doesn't exists
	if sect == nil {
		var err error

		sect, err = f.AddSection(section)
		if err != nil {
			f.Logger.Error("Failed to add new section", "section", section, "error", err)
			return false
		}
	}

	val := cast.ToString(value)
	isSet := false

	action := "Updating"
	if !f.HasKey(section, key) {
		action = "Adding"
	}

	if unique {
		sect.SetUniqueKey(key, val)
		_, isSet = sect.GetKey(key)
	} else {
		sect.SetKey(key, val)
		allKeys := sect.GetKeys(key)
		isSet = slices.Contains(allKeys, val)
	}

	f.Logger.Debug(fmt.Sprintf("%s key", action),
		"function", "setKeyValue", "section", section, "key", key, "value", val, "unique", unique)
	return isSet
}
