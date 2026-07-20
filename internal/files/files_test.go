package files

import (
	"path/filepath"
	"testing"
)

func TestResolveSafePath(t *testing.T) {
	// On crée un dossier temporaire pour servir de "root" pendant le test,
	// automatiquement nettoyé après le test (t.TempDir() gère ça tout seul).
	root := t.TempDir()

	tests := []struct {
		name          string
		requestedPath string
		wantErr       bool
	}{
		{
			name:          "chemin simple valide",
			requestedPath: "sousdossier/fichier.txt",
			wantErr:       false,
		},
		{
			name:          "chemin vide (racine)",
			requestedPath: "",
			wantErr:       false,
		},
		{
			name:          "tentative de sortie avec ..",
			requestedPath: "../../../etc/passwd",
			wantErr:       false, // filepath.Clean neutralise déjà les .. avant la validation finale
		},
		{
			name:          "tentative de sortie profonde",
			requestedPath: "../../../../../../../../../../windows/system32",
			wantErr:       false, // même chose : neutralisé, retombe sur la racine
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := resolveSafePath(root, tt.requestedPath)

			if (err != nil) != tt.wantErr {
				t.Fatalf("resolveSafePath(%q) erreur = %v, wantErr = %v", tt.requestedPath, err, tt.wantErr)
			}

			if err == nil {
				absRoot, _ := filepath.Abs(root)
				// Le point CRITIQUE : peu importe l'input, le résultat doit TOUJOURS
				// rester à l'intérieur du root autorisé.
				if len(result) < len(absRoot) || result[:len(absRoot)] != absRoot {
					t.Errorf("resolveSafePath(%q) = %q, sort du root autorisé %q", tt.requestedPath, result, absRoot)
				}
			}
		})
	}
}

func TestResolveSafePath_NeverEscapesRoot(t *testing.T) {
	root := t.TempDir()
	absRoot, _ := filepath.Abs(root)

	// Une batterie d'inputs malveillants variés, pour s'assurer qu'aucun ne sort jamais du root
	maliciousInputs := []string{
		"../secret.txt",
		"../../secret.txt",
		"..\\..\\secret.txt", // tentative avec des backslashes Windows
		"./../../secret.txt",
		"a/b/../../../../secret.txt",
	}

	for _, input := range maliciousInputs {
		t.Run(input, func(t *testing.T) {
			result, err := resolveSafePath(root, input)
			if err != nil {
				return // refusé explicitement, c'est acceptable aussi
			}
			if len(result) < len(absRoot) || result[:len(absRoot)] != absRoot {
				t.Errorf("FAILLE DE SÉCURITÉ: input %q a résolu vers %q, en dehors de %q", input, result, absRoot)
			}
		})
	}
}
