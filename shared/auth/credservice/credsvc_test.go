package credservice

import (
	"os"
	"regexp"
	"testing"
)

var (
	data = []byte(`{
		"commitAnalysisCredentials": [
		  {
			"credKey": "commit_analysis_user_4",
			"username": "user_4",
			"password": "password_4"
		  },
		  {
			"credKey": "commit_analysis_user_5",
			"username": "user_5",
			"password": "password_5"
		  }
		],
		"datapullCredentials": [
		  {
			"credKey": "datapull_user_1",
			"username": "user_1",
			"password": "password_1"
		  },
		  {
			"credKey": "datapull_user_2",
			"username": "user_2",
			"password": "password_2"
		  },
		  {
			"credKey": "datapull_user_3",
			"username": "user_3",
			"password": "password_3"
		  }
		]
	  }`)
)

func CreateAuthTokensFile(filePath string) error {
	// Create a new file with the given path
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write the sample data to the file
	if _, err := file.Write(data); err != nil {
		return err
	}

	return nil
}

func RemoveAuthTokensFile(filePath string) error {
	// Remove the file if it exists
	if _, err := os.Stat(filePath); err == nil {
		if err := os.Remove(filePath); err != nil {
			return err
		}
	}
	return nil
}
func RemoveAllAuthReleatedFiles(filePath string) error {
	if err := RemoveAuthTokensFile(filePath); err != nil {
		return err
	}
	if err := RemoveAuthTokensFile(filePath + ".lock"); err != nil {
		return err
	}

	re := regexp.MustCompile(`\.json$`)
	backupFilePath := re.ReplaceAllString(filePath, ".backup.json")
	if err := RemoveAuthTokensFile(backupFilePath); err != nil {
		return err
	}

	return nil
}

func TestNormalizeAndPersistCredentials(t *testing.T) {
	// Create a temporary file for testing
	filePath := "testdata/auth_tokens.json"
	if err := CreateAuthTokensFile(filePath); err != nil {
		t.Fatalf("failed to create auth_tokens.json file: %v", err)
	}
	defer func() {
		if err := RemoveAllAuthReleatedFiles(filePath); err != nil {
			t.Fatalf("failed to remove auth_tokens.json file: %v", err)
		}
	}()

	// Call the function to normalize and persist credentials
	if _, err := NormalizeAndPersistCredentials(filePath); err != nil {
		t.Fatalf("failed to normalize and persist credentials: %v", err)
	}
}

func TestLoadAuthTokensFromFileAndValidate(t *testing.T) {
	// Create a temporary file for testing
	filePath := "testdata/auth_tokens.json"
	if err := CreateAuthTokensFile(filePath); err != nil {
		t.Fatalf("failed to create auth_tokens.json file: %v", err)
	}
	defer func() {
		if err := RemoveAllAuthReleatedFiles(filePath); err != nil {
			t.Fatalf("failed to remove auth_tokens.json file: %v", err)
		}
	}()

	_, _, err := LoadAuthTokensFromFileAndValidate("testdata/auth_tokens.json")
	if err != nil {
		t.Fatalf("failed to load and validate credentials: %v", err)
	}
}
