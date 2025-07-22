package ub2fa

import (
	"fmt"
	"image/png"
	"io"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

type TotpService interface {
	GenerateTotp(accountName string) (string, error)
	ValidateTotp(url string, code string) (bool, error)
	GenerateTotpPng(w io.Writer, url string) error
	GetTotpCode(url string) (string, error)
}

type TotpServiceImpl struct {
	issuer string
}

func NewTotpService(issuer string) *TotpServiceImpl {
	return &TotpServiceImpl{
		issuer: issuer,
	}
}

func (s *TotpServiceImpl) GenerateTotp(accountName string) (string, error) {
	opts := totp.GenerateOpts{
		Issuer:      s.issuer,
		AccountName: accountName,
	}

	key, err := totp.Generate(opts)
	if err != nil {
		return "", fmt.Errorf("generating totp - failed to generate totp key: %w", err)
	}
	return key.URL(), nil
}

func (s *TotpServiceImpl) GetTotpCode(url string) (string, error) {
	key, err := otp.NewKeyFromURL(url)
	if err != nil {
		return "", fmt.Errorf("validating totp - failed to generate totp key: %w", err)
	}
	return totp.GenerateCode(key.Secret(), time.Now())
}

func (s *TotpServiceImpl) ValidateTotp(url string, code string) (bool, error) {
	key, err := otp.NewKeyFromURL(url)
	if err != nil {
		return false, fmt.Errorf("validating totp - failed to generate totp key: %w", err)
	}

	return totp.Validate(code, key.Secret()), nil
}

func (s *TotpServiceImpl) GenerateTotpPng(w io.Writer, url string) error {
	key, err := otp.NewKeyFromURL(url)

	if err != nil {
		return fmt.Errorf("generating png - failed to generate totp key: %w", err)
	}
	image, err := key.Image(200, 200)
	if err != nil {
		return fmt.Errorf("generating png - failed to generate image: %w", err)
	}
	err = png.Encode(w, image)
	if err != nil {
		return fmt.Errorf("generating png - failed to encode image: %w", err)
	}
	return nil
}
