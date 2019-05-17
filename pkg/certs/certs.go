package certs

import "io/ioutil"

type Certs struct {
	TrustedCa   string
	Certificate string
	PrivateKey  string
}

func LoadCerts(directory string) (*Certs, error) {
	trustedCa, err := ioutil.ReadFile(directory + "/ca.crt")
	if err != nil {
		return nil, err
	}

	certificate, err := ioutil.ReadFile(directory + "/client.crt")
	if err != nil {
		return nil, err
	}

	privateKey, err := ioutil.ReadFile(directory + "/client.key")
	if err != nil {
		return nil, err
	}

	return &Certs{
		TrustedCa:   string(trustedCa),
		Certificate: string(certificate),
		PrivateKey:  string(privateKey),
	}, nil
}
