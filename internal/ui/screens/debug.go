package screens

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"flagged-it/internal/ui/components"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

type DebugScreen struct {
	content         *fyne.Container
	backFunc        func()
	window          fyne.Window
	dataDir         string
	fileNames       []string
	currentFile     string
	data            []map[string]interface{}
	fileSelect      *widget.Select
	elementSelect   *widget.Select
	editorContainer *fyne.Container
}

func NewDebugScreen(backFunc func(), window fyne.Window) *DebugScreen {
	d := &DebugScreen{
		backFunc: backFunc,
		window:   window,
		dataDir:  "internal/data/sources",
	}
	d.loadFileList()
	d.setupUI()
	return d
}

func (d *DebugScreen) loadFileList() {
	files, err := os.ReadDir(d.dataDir)
	if err != nil {
		return
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".json") {
			d.fileNames = append(d.fileNames, file.Name())
		}
	}
}

func (d *DebugScreen) setupUI() {
	title := widget.NewLabel("Debug Data Editor")
	title.TextStyle = fyne.TextStyle{Bold: true}

	d.fileSelect = widget.NewSelect(d.fileNames, d.onFileSelected)
	d.fileSelect.PlaceHolder = "Select a file..."

	d.elementSelect = widget.NewSelect([]string{}, d.onElementSelected)
	d.elementSelect.PlaceHolder = "Select an element..."
	d.elementSelect.Disable()

	d.editorContainer = container.NewVBox()

	saveBtn := components.NewButton("Save Changes", d.saveFile)
	addBtn := components.NewButton("Add Element", d.addElement)
	removeBtn := components.NewButton("Remove Element", d.removeElement)
	backBtn := components.NewButton("Back to Dashboard", d.backFunc)

	header := container.NewVBox(
		title,
		widget.NewSeparator(),
		widget.NewLabel("1. Select Data File:"),
		container.NewGridWithColumns(2, d.fileSelect, widget.NewLabel("")),
		widget.NewSeparator(),
		widget.NewLabel("2. Select Element:"),
		container.NewGridWithColumns(2, container.NewVBox(d.elementSelect, container.NewHBox(addBtn, removeBtn)), widget.NewLabel("")),
		widget.NewSeparator(),
	)

	footer := container.NewVBox(
		widget.NewSeparator(),
		container.NewHBox(saveBtn, backBtn),
	)

	d.content = container.NewBorder(header, footer, nil, nil, container.NewScroll(d.editorContainer))
}

func (d *DebugScreen) onFileSelected(fileName string) {
	d.loadFile(fileName)
}

func (d *DebugScreen) loadFile(fileName string) {
	filePath := filepath.Join(d.dataDir, fileName)
	d.currentFile = filePath

	content, err := os.ReadFile(filePath)
	if err != nil {
		dialog.ShowError(fmt.Errorf("Failed to read file: %v", err), d.window)
		return
	}

	var rawData interface{}
	if err := json.Unmarshal(content, &rawData); err != nil {
		dialog.ShowError(fmt.Errorf("Invalid JSON: %v", err), d.window)
		return
	}

	d.data = d.normalizeData(rawData)
	d.updateElementList()
	d.elementSelect.Enable()
	d.elementSelect.ClearSelected()
	d.editorContainer.RemoveAll()
	d.editorContainer.Refresh()
}

func (d *DebugScreen) updateElementList() {
	var elements []string
	for i, item := range d.data {
		name := d.getCountryName(item, i)
		elements = append(elements, name)
	}
	sort.Strings(elements)
	d.elementSelect.Options = elements
	d.elementSelect.SetSelected("")
	d.elementSelect.Refresh()
}

func (d *DebugScreen) normalizeData(rawData interface{}) []map[string]interface{} {
	if arr, ok := rawData.([]interface{}); ok {
		result := make([]map[string]interface{}, len(arr))
		for i, item := range arr {
			if m, ok := item.(map[string]interface{}); ok {
				result[i] = m
			}
		}
		return result
	}

	if obj, ok := rawData.(map[string]interface{}); ok {
		if features, ok := obj["features"].([]interface{}); ok {
			result := make([]map[string]interface{}, len(features))
			for i, item := range features {
				if m, ok := item.(map[string]interface{}); ok {
					result[i] = m
				}
			}
			return result
		}

		var result []map[string]interface{}
		for key, value := range obj {
			if m, ok := value.(map[string]interface{}); ok {
				m["_key"] = key
				result = append(result, m)
			}
		}
		return result
	}

	return []map[string]interface{}{}
}

func (d *DebugScreen) getCountryName(item map[string]interface{}, index int) string {
	if key, ok := item["_key"].(string); ok {
		if name, ok := item["name"].(string); ok {
			return fmt.Sprintf("%s (%s)", name, key)
		}
		return key
	}

	if props, ok := item["properties"].(map[string]interface{}); ok {
		if name, ok := props["name"].(string); ok {
			return name
		}
	}

	if nameVal, ok := item["name"]; ok {
		if nameMap, ok := nameVal.(map[string]interface{}); ok {
			if common, ok := nameMap["common"].(string); ok {
				return common
			}
		}
		if nameStr, ok := nameVal.(string); ok {
			return nameStr
		}
	}

	return fmt.Sprintf("Element %d", index+1)
}

func (d *DebugScreen) onElementSelected(element string) {
	if element == "" {
		return
	}

	for i, item := range d.data {
		if d.getCountryName(item, i) == element {
			d.showElementEditor(i)
			return
		}
	}
}

func (d *DebugScreen) showElementEditor(index int) {
	if index >= len(d.data) {
		return
	}

	d.editorContainer.RemoveAll()

	addParamBtn := components.NewButton("+ Add Parameter", func() {
		d.addParameter(index)
	})
	d.editorContainer.Add(addParamBtn)
	d.editorContainer.Add(widget.NewSeparator())

	item := d.data[index]
	for key := range item {
		value := item[key]
		jsonBytes, _ := json.MarshalIndent(value, "", "  ")
		valStr := string(jsonBytes)

		keyEntry := widget.NewEntry()
		keyEntry.SetText(key)
		capturedKey := key
		capturedIndex := index
		keyEntry.OnChanged = func(newKey string) {
			if newKey != "" && newKey != capturedKey {
				d.renameParameter(capturedIndex, capturedKey, newKey)
				capturedKey = newKey
			}
		}

		removeBtn := components.NewButton("Remove", func() {
			d.removeParameter(capturedIndex, capturedKey)
		})

		headerRow := container.NewBorder(nil, nil, nil, removeBtn, keyEntry)

		entry := widget.NewMultiLineEntry()
		entry.SetText(valStr)
		entry.Wrapping = fyne.TextWrapWord
		entry.SetMinRowsVisible(5)

		entry.OnChanged = func(s string) {
			d.updateValue(capturedIndex, capturedKey, s)
		}

		paramContainer := container.NewVBox(headerRow, entry)
		fieldContainer := container.NewGridWithColumns(2, paramContainer, widget.NewLabel(""))

		d.editorContainer.Add(fieldContainer)
		d.editorContainer.Add(widget.NewSeparator())
	}

	d.editorContainer.Refresh()
}

func (d *DebugScreen) updateValue(index int, key, value string) {
	if index >= len(d.data) {
		return
	}

	var val interface{}
	if err := json.Unmarshal([]byte(value), &val); err == nil {
		d.data[index][key] = val
	} else {
		d.data[index][key] = value
	}

	if key == "name" || key == "_key" {
		d.updateElementList()
	}
}

func (d *DebugScreen) renameParameter(index int, oldKey, newKey string) {
	if index >= len(d.data) || oldKey == newKey {
		return
	}

	if _, exists := d.data[index][newKey]; exists {
		dialog.ShowError(fmt.Errorf("Parameter '%s' already exists", newKey), d.window)
		return
	}

	d.data[index][newKey] = d.data[index][oldKey]
	delete(d.data[index], oldKey)

	if oldKey == "name" || oldKey == "_key" || newKey == "name" || newKey == "_key" {
		d.updateElementList()
	}
}

func (d *DebugScreen) addParameter(index int) {
	if index >= len(d.data) {
		return
	}

	newKey := "new_parameter"
	counter := 1
	for {
		if _, exists := d.data[index][newKey]; !exists {
			break
		}
		newKey = fmt.Sprintf("new_parameter_%d", counter)
		counter++
	}

	d.data[index][newKey] = ""
	d.showElementEditor(index)
}

func (d *DebugScreen) removeParameter(index int, key string) {
	if index >= len(d.data) {
		return
	}

	dialog.ShowConfirm("Confirm Delete",
		fmt.Sprintf("Remove parameter '%s'?", key),
		func(confirmed bool) {
			if confirmed {
				delete(d.data[index], key)
				d.showElementEditor(index)
				if key == "name" || key == "_key" {
					d.updateElementList()
				}
			}
		}, d.window)
}

func (d *DebugScreen) saveFile() {
	var outputData interface{}
	if len(d.data) > 0 {
		if _, hasKey := d.data[0]["_key"]; hasKey {
			obj := make(map[string]interface{})
			for _, item := range d.data {
				key := item["_key"].(string)
				copy := make(map[string]interface{})
				for k, v := range item {
					if k != "_key" {
						copy[k] = v
					}
				}
				obj[key] = copy
			}
			outputData = obj
		} else if _, hasProps := d.data[0]["properties"]; hasProps {
			outputData = map[string]interface{}{
				"type":     "FeatureCollection",
				"features": d.data,
			}
		} else {
			outputData = d.data
		}
	} else {
		outputData = d.data
	}

	jsonData, err := json.MarshalIndent(outputData, "", "    ")
	if err != nil {
		dialog.ShowError(fmt.Errorf("Failed to marshal JSON: %v", err), d.window)
		return
	}

	if err := os.WriteFile(d.currentFile, jsonData, 0644); err != nil {
		dialog.ShowError(fmt.Errorf("Failed to save file: %v", err), d.window)
		return
	}

	dialog.ShowInformation("Success", "File saved successfully!", d.window)
}

func (d *DebugScreen) addElement() {
	if d.currentFile == "" {
		dialog.ShowError(fmt.Errorf("Please select a file first"), d.window)
		return
	}

	newElement := make(map[string]interface{})
	if len(d.data) > 0 {
		if _, hasKey := d.data[0]["_key"]; hasKey {
			newElement["_key"] = "NEW"
			newElement["name"] = "New Element"
		} else if _, hasProps := d.data[0]["properties"]; hasProps {
			newElement["type"] = "Feature"
			newElement["properties"] = map[string]interface{}{"name": "New Element"}
		} else {
			newElement["name"] = map[string]interface{}{"common": "New Element"}
		}
	} else {
		newElement["name"] = "New Element"
	}

	d.data = append(d.data, newElement)
	d.updateElementList()
	d.elementSelect.SetSelected(d.getCountryName(newElement, len(d.data)-1))
}

func (d *DebugScreen) removeElement() {
	selected := d.elementSelect.Selected
	if selected == "" {
		dialog.ShowError(fmt.Errorf("Please select an element to remove"), d.window)
		return
	}

	for i, item := range d.data {
		if d.getCountryName(item, i) == selected {
			dialog.ShowConfirm("Confirm Delete",
				fmt.Sprintf("Are you sure you want to delete '%s'?", selected),
				func(confirmed bool) {
					if confirmed {
						d.data = append(d.data[:i], d.data[i+1:]...)
						d.updateElementList()
						d.elementSelect.ClearSelected()
						d.editorContainer.RemoveAll()
						d.editorContainer.Refresh()
					}
				}, d.window)
			return
		}
	}
}

func (d *DebugScreen) GetContent() *fyne.Container {
	return d.content
}
