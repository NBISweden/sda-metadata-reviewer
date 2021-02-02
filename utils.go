package main

func getAccessionIdsAndSchemas(metadataCollections []MetadataCollection) ([]string, []string) {

	var schemas []string
	var accessionIds []string

	for _, col := range metadataCollections {
		for _, obj := range col.MetadataObjects {
			accessionIds = append(accessionIds, obj.AccessionID)
			schemas = append(schemas, obj.Schema)
		}
	}
	return removeStrDuplicates(accessionIds), removeStrDuplicates(schemas)
}

func removeStrDuplicates(elements []string) []string {
	encountered := map[string]bool{}
	res := []string{}

	for v := range elements {
		if encountered[elements[v]] == true {
		} else {
			encountered[elements[v]] = true
			res = append(res, elements[v])
		}
	}
	return res
}
