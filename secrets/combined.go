package secrets

import "fmt"

type combinedSupplier struct {
	suppliers []SecretSupplier
}

func NullSupplier() SecretSupplier {
	return &combinedSupplier{}
}

func CombineSuppliers(first SecretSupplier, last SecretSupplier) SecretSupplier {
	cs_first, ok_first := first.(*combinedSupplier)
	cs_last, ok_last := last.(*combinedSupplier)
	var suppliers []SecretSupplier
	if ok_first && ok_last {
		suppliers = append(cs_last.suppliers, cs_first.suppliers...)
	} else if ok_first {
		suppliers = append([]SecretSupplier{last}, cs_first.suppliers...)
	} else if ok_last {
		suppliers = append(cs_last.suppliers, first)
	} else {
		suppliers = []SecretSupplier{last, first}
	}
	if len(suppliers) == 1 {
		return suppliers[0]
	} else {
		return &combinedSupplier{suppliers: suppliers}
	}
}

func (s *combinedSupplier) GetSecret(key string) (string, error) {
	for _, supplier := range s.suppliers {
		secret, err := supplier.GetSecret(key)
		if err == nil {
			return secret, nil
		}
	}
	return "", fmt.Errorf("no secret found for key '%s'", key)
}
