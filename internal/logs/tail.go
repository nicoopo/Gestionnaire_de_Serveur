package logs

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"time"
)

// TailFile lit un fichier depuis sa fin et pousse chaque nouvelle ligne dans le channel `lines`,
// jusqu'à ce que le context soit annulé.
func TailFile(ctx context.Context, path string, lines chan<- string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("impossible d'ouvrir le fichier de log: %w", err)
	}
	defer file.Close()

	// On se positionne à la fin du fichier pour ne lire que les nouvelles lignes
	if _, err := file.Seek(0, os.SEEK_END); err != nil {
		return fmt.Errorf("impossible de se positionner en fin de fichier: %w", err)
	}

	reader := bufio.NewReader(file)
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil

		case <-ticker.C:
			for {
				line, err := reader.ReadString('\n')
				if len(line) > 0 {
					select {
					case lines <- line:
					case <-ctx.Done():
						return nil
					}
				}
				if err != nil {
					break // pas de nouvelle ligne pour l'instant, on retente au prochain tick
				}
			}
		}
	}
}
