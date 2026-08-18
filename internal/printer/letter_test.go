package printer_test

import (
	"testing"

	"github.com/jpporta/ticket-control/internal/printer"
)

func TestRenderLetter(t *testing.T) {
	input := printer.LetterInput{
		Title:    "Carta de Teste",
		Date:     "17 de Agosto de 2026",
		To:       "Amigo",
		ToLabel:  "Para",
		From:     "Autor",
		SignOff:  "Atenciosamente,",
		Font:     "Libertinus Serif",
		FontSize: "11pt",
		Justify:  true,
		Content: `# Cabeçalho

Este é um parágrafo de teste com formatação em **negrito** e _itálico_.

- Ponto 1
- Ponto 2
`,
	}

	img, err := printer.RenderLetter(input)
	if err != nil {
		t.Fatalf("RenderLetter failed: %v", err)
	}
	if img == nil {
		t.Fatal("expected non-nil image")
	}
	b := img.Bounds()
	if b.Max.Y%8 != 0 {
		t.Errorf("image height %d is not a multiple of 8", b.Max.Y)
	}
}
