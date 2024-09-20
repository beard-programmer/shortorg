package core

type DestinationURL = Url

func DestinationURLFromString(s string) (*DestinationURL, error) {
	url, err := NewURL(s)
	if err != nil {
		return nil, err
	}

	return url, nil
}
