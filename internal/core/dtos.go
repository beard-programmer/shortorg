package core

type URLDto struct {
	Value string
}

func (u *URL) IntoDto() URLDto {
	return URLDto{u.String()}
}
func (dto URLDto) IntoDomain() (*URL, error) {
	return NewURL(dto.Value)
}

type LinkSlugDto struct {
	Value string
}

func (s LinkSlug) IntoDto() LinkSlugDto {
	return LinkSlugDto{Value: s.Value()}
}

func (dto LinkSlugDto) IntoDomain() (*LinkSlug, error) {
	return NewLinkSlug(dto.Value)
}

type LinkHostDto struct {
	Hostname  string
	IsBranded bool
}

func (h LinkHost) IntoDto() LinkHostDto {
	return LinkHostDto{Hostname: h.Hostname(), IsBranded: h.IsBranded()}
}

func (dto LinkHostDto) IntoDomain() (*LinkHost, error) {
	return NewLinkHost(&dto.Hostname)
}

type LinkKeyDto struct {
	Value int64
}

func (k LinkKey) IntoDto() LinkKeyDto {
	return LinkKeyDto{Value: k.Value()}
}

func (dto LinkKeyDto) IntoDomain() (*LinkKey, error) {
	return NewLinkKey(dto.Value)
}
