package outline

// RemoveDuplicates removes duplicate entries from the outline
func (o *Outline) RemoveDuplicates() {
	// Remove duplicate functions
	funcSet := make(map[string]bool)
	var uniqueFuncs []string
	for _, f := range o.Funcs {
		if !funcSet[f] {
			funcSet[f] = true
			uniqueFuncs = append(uniqueFuncs, f)
		}
	}
	o.Funcs = uniqueFuncs

	// Remove duplicate vars
	varSet := make(map[string]bool)
	var uniqueVars []string
	for _, v := range o.Vars {
		if !varSet[v] {
			varSet[v] = true
			uniqueVars = append(uniqueVars, v)
		}
	}
	o.Vars = uniqueVars

	// Remove duplicates from file info
	for _, fileInfo := range o.Files {
		// Remove duplicate functions in file
		fileFuncSet := make(map[string]bool)
		var uniqueFileFuncs []FunctionInfo
		for _, f := range fileInfo.Functions {
			if !fileFuncSet[f.Name] {
				fileFuncSet[f.Name] = true
				uniqueFileFuncs = append(uniqueFileFuncs, f)
			}
		}
		fileInfo.Functions = uniqueFileFuncs

		// Remove duplicate types in file
		typeSet := make(map[string]bool)
		var uniqueTypes []string
		for _, t := range fileInfo.Types {
			if !typeSet[t] {
				typeSet[t] = true
				uniqueTypes = append(uniqueTypes, t)
			}
		}
		fileInfo.Types = uniqueTypes

		// Remove duplicate vars in file
		fileVarSet := make(map[string]bool)
		var uniqueFileVars []string
		for _, v := range fileInfo.Vars {
			if !fileVarSet[v] {
				fileVarSet[v] = true
				uniqueFileVars = append(uniqueFileVars, v)
			}
		}
		fileInfo.Vars = uniqueFileVars
	}
}
