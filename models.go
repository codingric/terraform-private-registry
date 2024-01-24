package main

type VersionJson struct {
	Version []Version `json:"versions"`
}

type Version struct {
	Version   string     `json:"version"`
	Protocols []string   `json:"protocols"`
	Platforms []Platform `json:"platforms"`
}

type Platform struct {
	Os                  string     `json:"os"`
	Arch                string     `json:"arch"`
	Filename            string     `json:"filename"`
	DownloadUrl         string     `json:"download_url"`
	ShasumsUrl          string     `json:"shasums_url"`
	ShasumsSignatureUrl string     `json:"shasums_signature_url"`
	Shasum              string     `json:"shasum"`
	SigningKey          SigningKey `json:"signing_keys"`
}

type SigningKey struct {
	GpgPublicKey []GpgPublicKey `json:"gpg_public_keys"`
}

type GpgPublicKey struct {
	KeyId          string `json:"key_id"`
	AsciiArmor     string `json:"ascii_armor"`
	TrustSignature string `json:"trust_signature"`
	Source         string `json:"source"`
	SourceUrl      string `json:"source_url"`
}
