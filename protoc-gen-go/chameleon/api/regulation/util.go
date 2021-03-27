package regulation

import "strings"

func GetCrowdType(crowd string) CrowdType {
	crowd = strings.ToLower(crowd)

	for _, key := range []string{"women", "woman", "female"} {
		if strings.Contains(crowd, key) {
			return CrowdType_CrowdTypeWomen
		}
	}

	for _, key := range []string{"men", "man", "male"} {
		if strings.Contains(crowd, key) {
			return CrowdType_CrowdTypeMen
		}
	}

	for _, key := range []string{"kid", "child", "boy", "girl"} {
		if strings.Contains(crowd, key) {
			return CrowdType_CrowdTypeKids
		}
	}
	return CrowdType_CrowdTypeAny
}

func GetCrowdName(ct CrowdType) string {
	switch ct {
	case CrowdType_CrowdTypeMen:
		return "men"
	case CrowdType_CrowdTypeWomen:
		return "women"
	case CrowdType_CrowdTypeKids:
		return "kids"
	}
	return ""
}
