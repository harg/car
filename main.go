package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func createFakeXLSArchive(archiveName string, files []string) error {
	out, err := os.Create(archiveName)
	if err != nil {
		return err
	}
	defer out.Close()
	// Write fake XLS header (BIFF8 magic number: D0 CF 11 E0 A1 B1 1A E1)
	fakeXLSHeader := []byte{0xD0, 0xCF, 0x11, 0xE0, 0xA1, 0xB1, 0x1A, 0xE1}
	if _, err := out.Write(fakeXLSHeader); err != nil {
		return err
	}
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			return err
		}
		filename := filepath.Base(file)
		nameLen := int32(len(filename))
		fileSize := int64(len(data))
		if err := binary.Write(out, binary.LittleEndian, nameLen); err != nil {
			return err
		}
		if _, err := out.Write([]byte(filename)); err != nil {
			return err
		}
		if err := binary.Write(out, binary.LittleEndian, fileSize); err != nil {
			return err
		}
		if _, err := out.Write(data); err != nil {
			return err
		}
	}
	return nil
}

func extractFakeXLSArchive(archiveName string, outputDir string) error {
	in, err := os.Open(archiveName)
	if err != nil {
		return err
	}
	defer in.Close()
	// Skip fake XLS header
	if _, err := in.Seek(8, io.SeekStart); err != nil {
		return err
	}
	for {
		var nameLen int32
		if err := binary.Read(in, binary.LittleEndian, &nameLen); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		nameBytes := make([]byte, nameLen)
		if _, err := io.ReadFull(in, nameBytes); err != nil {
			return err
		}
		filename := string(nameBytes)
		var fileSize int64
		if err := binary.Read(in, binary.LittleEndian, &fileSize); err != nil {
			return err
		}
		data := make([]byte, fileSize)
		if _, err := io.ReadFull(in, data); err != nil {
			return err
		}
		outPath := filepath.Join(outputDir, filename)
		if err := os.WriteFile(outPath, data, 0644); err != nil {
			return err
		}
	}
	return nil
}

// ensureXLSExtension adds .xls extension if not present
func ensureXLSExtension(filename string) string {
	if !strings.HasSuffix(strings.ToLower(filename), ".xls") {
		return filename + ".xls"
	}
	return filename
}

// promptOverwrite asks user if they want to overwrite existing file
func promptOverwrite(filename string) bool {
	fmt.Printf("File '%s' already exists. Overwrite? (y/N): ", filename)

	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false // Default to NO on error
	}

	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes"
}

// checkFileExists checks if file exists and prompts for overwrite
func checkFileExists(filename string) (bool, error) {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return true, nil // File doesn't exist, safe to proceed
	} else if err != nil {
		return false, err // Some other error
	}

	// File exists, ask user
	if promptOverwrite(filename) {
		return true, nil
	}

	return false, nil // User chose not to overwrite
}

func main() {
	if len(os.Args) < 4 {
		fmt.Println("Usage:")
		fmt.Println("  car pack archive.xls file1 file2 ...")
		fmt.Println("  car unpack archive.xls output_dir")
		return
	}
	command := os.Args[1]
	switch command {
	case "pack":
		archiveName := os.Args[2]
		files := os.Args[3:]

		// Ensure .xls extension
		archiveName = ensureXLSExtension(archiveName)

		// Check if file exists and prompt for overwrite
		canProceed, err := checkFileExists(archiveName)
		if err != nil {
			fmt.Printf("Error checking file existence: %v\n", err)
			return
		}

		if !canProceed {
			fmt.Println("Operation cancelled.")
			return
		}

		if err := createFakeXLSArchive(archiveName, files); err != nil {
			fmt.Println("Error packing files: ", err)
		} else {
			fmt.Println("Archive created: ", archiveName)
		}
	case "unpack":
		archiveName := os.Args[2]
		outputDir := os.Args[3]
		if err := extractFakeXLSArchive(archiveName, outputDir); err != nil {
			fmt.Println("Error extracting archive: ", err)
		} else {
			fmt.Println("Archive extracted to: ", outputDir)
		}
	default:
		fmt.Println("Unknown command: ", command)
	}
}
