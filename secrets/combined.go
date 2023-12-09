package secrets

import "fmt"

type combinedSupplier struct {
	suppliers []SecretSupplier
}

func NullSupplier() *combinedSupplier {
	return &combinedSupplier{}
}

func CombineSuppliers(a SecretSupplier, b SecretSupplier) *combinedSupplier {
	csa, oka := a.(*combinedSupplier)
	csb, okb := b.(*combinedSupplier)
	if oka && okb {
		return &combinedSupplier{suppliers: append(csa.suppliers, csb.suppliers...)}
	}
	if oka {
		return &combinedSupplier{suppliers: append(csa.suppliers, b)}
	}
	if okb {
		return &combinedSupplier{suppliers: append([]SecretSupplier{a}, csb.suppliers...)}
	}
	return &combinedSupplier{suppliers: []SecretSupplier{a, b}}
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
