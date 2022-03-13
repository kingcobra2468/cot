package gvoice

// Link binds a given GVoice number with a client number for
// fetching future commands.
type Link struct {
	GVoiceNumber string
	ClientNumber string
}

func (l Link) Messages() []string {
	return []string{"test"}
}
