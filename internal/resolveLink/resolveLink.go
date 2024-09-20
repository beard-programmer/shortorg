package resolveLink

type resolveLinkUseCase struct {
}

type LinkWasResolved struct {
	destination string
}

func (u *resolveLinkUseCase) ResolveLink(link string) (string, error) {

}
