package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/quick"
)

// TestProperty_EntryAdditionAccuracy tests Property 2: エントリ追加の正確性
// **Validates: Requirements 2.10, 4.10**
//
// Property: For any valid entry (type, name, definition), when syntax check succeeds,
// adding that entry to Management_File means the entry will exist when the file is reloaded.
func TestProperty_EntryAdditionAccuracy(t *testing.T) {
	// プロパティ関数: エントリを追加して再読み込みすると、そのエントリが存在する
	property := func(entryType string, name string, definition string) bool {
		// 入力を正規化して有効な値にする
		entryType = normalizeEntryType(entryType)
		name = normalizeName(name)
		definition = normalizeDefinition(definition)

		// 空白のみの入力はスキップ（これは別のプロパティでテストされる）
		if strings.TrimSpace(name) == "" || strings.TrimSpace(definition) == "" {
			return true
		}

		// テスト用の一時ディレクトリを作成
		tempDir, err := os.MkdirTemp("", "fish-config-test-*")
		if err != nil {
			t.Logf("Failed to create temp dir: %v", err)
			return false
		}
		defer os.RemoveAll(tempDir)

		// テスト用のConfigManagerを作成
		testFilePath := filepath.Join(tempDir, "fish-configurator.fish")
		manager := &DefaultConfigManager{
			filePath: testFilePath,
			parser:   NewParser(),
		}

		// エントリを追加
		err = manager.AddEntry(entryType, name, definition)
		if err != nil {
			t.Logf("Failed to add entry: %v", err)
			return false
		}

		// ファイルを再読み込み
		config, err := manager.Load()
		if err != nil {
			t.Logf("Failed to reload config: %v", err)
			return false
		}

		// 追加したエントリが存在することを確認
		found := false
		for _, entry := range config.Entries {
			if entry.Type == entryType && entry.Name == name && entry.Definition == definition {
				found = true
				break
			}
		}

		if !found {
			t.Logf("Entry not found after reload: type=%s, name=%s, definition=%s", entryType, name, definition)
		}

		return found
	}

	// プロパティテストを実行
	config := &quick.Config{
		MaxCount: 100, // 100回のランダムテストを実行
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property violated: %v", err)
	}
}

// normalizeEntryType は entryType を "alias" または "abbr" に正規化する
func normalizeEntryType(s string) string {
	// 文字列の最初の文字を使って決定
	if len(s) == 0 || s[0]%2 == 0 {
		return "alias"
	}
	return "abbr"
}

// normalizeName は name を有効な識別子に正規化する
func normalizeName(s string) string {
	if len(s) == 0 {
		return "test"
	}
	
	// 英数字とアンダースコアのみを保持
	var builder strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			builder.WriteRune(r)
		}
	}
	
	result := builder.String()
	if result == "" {
		return "test"
	}
	
	// 最大長を制限
	if len(result) > 20 {
		result = result[:20]
	}
	
	return result
}

// normalizeDefinition は definition を有効なコマンドに正規化する
func normalizeDefinition(s string) string {
	if len(s) == 0 {
		return "echo test"
	}
	
	// シングルクォートをエスケープ（fish shellの構文に合わせる）
	s = strings.ReplaceAll(s, "'", "\\'")
	
	// 改行を削除
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	
	// 最大長を制限
	if len(s) > 50 {
		s = s[:50]
	}
	
	// 空白のみの場合はデフォルト値を返す
	if strings.TrimSpace(s) == "" {
		return "echo test"
	}
	
	return s
}

// TestProperty_ExistingEntryInvariance tests Property 3: 既存エントリの不変性
// **Validates: Requirements 2.11, 3.6**
//
// Property: For any Management_File and new entry, when adding or removing an entry,
// all entries other than the target remain unchanged.
func TestProperty_ExistingEntryInvariance(t *testing.T) {
	t.Run("AddEntry preserves existing entries", func(t *testing.T) {
		// プロパティ関数: エントリを追加しても既存のエントリは変更されない
		property := func(existingCount uint8, newEntryType string, newName string, newDefinition string) bool {
			// 既存エントリの数を制限（1-5個）
			count := int(existingCount%5) + 1

			// 新しいエントリの値を正規化
			newEntryType = normalizeEntryType(newEntryType)
			newName = normalizeName(newName)
			newDefinition = normalizeDefinition(newDefinition)

			// 空白のみの入力はスキップ
			if strings.TrimSpace(newName) == "" || strings.TrimSpace(newDefinition) == "" {
				return true
			}

			// テスト用の一時ディレクトリを作成
			tempDir, err := os.MkdirTemp("", "fish-config-test-*")
			if err != nil {
				t.Logf("Failed to create temp dir: %v", err)
				return false
			}
			defer os.RemoveAll(tempDir)

			// テスト用のConfigManagerを作成
			testFilePath := filepath.Join(tempDir, "fish-configurator.fish")
			manager := &DefaultConfigManager{
				filePath: testFilePath,
				parser:   NewParser(),
			}

			// 既存のエントリを追加
			existingEntries := make([]Entry, count)
			for i := 0; i < count; i++ {
				entryType := "alias"
				if i%2 == 1 {
					entryType = "abbr"
				}
				name := fmt.Sprintf("existing%d", i)
				definition := fmt.Sprintf("echo existing%d", i)

				existingEntries[i] = Entry{
					Type:       entryType,
					Name:       name,
					Definition: definition,
				}

				err := manager.AddEntry(entryType, name, definition)
				if err != nil {
					t.Logf("Failed to add existing entry: %v", err)
					return false
				}
			}

			// 新しいエントリを追加
			err = manager.AddEntry(newEntryType, newName, newDefinition)
			if err != nil {
				t.Logf("Failed to add new entry: %v", err)
				return false
			}

			// ファイルを再読み込み
			config, err := manager.Load()
			if err != nil {
				t.Logf("Failed to reload config: %v", err)
				return false
			}

			// 既存のエントリがすべて保持されていることを確認
			for _, existingEntry := range existingEntries {
				found := false
				for _, entry := range config.Entries {
					if entry.Type == existingEntry.Type &&
						entry.Name == existingEntry.Name &&
						entry.Definition == existingEntry.Definition {
						found = true
						break
					}
				}
				if !found {
					t.Logf("Existing entry not preserved: type=%s, name=%s, definition=%s",
						existingEntry.Type, existingEntry.Name, existingEntry.Definition)
					return false
				}
			}

			return true
		}

		// プロパティテストを実行
		config := &quick.Config{
			MaxCount: 50, // 50回のランダムテストを実行
		}

		if err := quick.Check(property, config); err != nil {
			t.Errorf("Property violated: %v", err)
		}
	})

	t.Run("RemoveEntry preserves other entries", func(t *testing.T) {
		// プロパティ関数: エントリを削除しても他のエントリは変更されない
		property := func(totalCount uint8, removeIndex uint8) bool {
			// エントリの総数を制限（2-6個、削除するので最低2個必要）
			count := int(totalCount%5) + 2
			// 削除するインデックスを範囲内に制限
			removeIdx := int(removeIndex) % count

			// テスト用の一時ディレクトリを作成
			tempDir, err := os.MkdirTemp("", "fish-config-test-*")
			if err != nil {
				t.Logf("Failed to create temp dir: %v", err)
				return false
			}
			defer os.RemoveAll(tempDir)

			// テスト用のConfigManagerを作成
			testFilePath := filepath.Join(tempDir, "fish-configurator.fish")
			manager := &DefaultConfigManager{
				filePath: testFilePath,
				parser:   NewParser(),
			}

			// エントリを追加
			allEntries := make([]Entry, count)
			for i := 0; i < count; i++ {
				entryType := "alias"
				if i%2 == 1 {
					entryType = "abbr"
				}
				name := fmt.Sprintf("entry%d", i)
				definition := fmt.Sprintf("echo entry%d", i)

				allEntries[i] = Entry{
					Type:       entryType,
					Name:       name,
					Definition: definition,
				}

				err := manager.AddEntry(entryType, name, definition)
				if err != nil {
					t.Logf("Failed to add entry: %v", err)
					return false
				}
			}

			// 1つのエントリを削除
			toRemove := allEntries[removeIdx]
			err = manager.RemoveEntry(toRemove.Type, toRemove.Name)
			if err != nil {
				t.Logf("Failed to remove entry: %v", err)
				return false
			}

			// ファイルを再読み込み
			config, err := manager.Load()
			if err != nil {
				t.Logf("Failed to reload config: %v", err)
				return false
			}

			// 削除されたエントリ以外のすべてのエントリが保持されていることを確認
			for i, entry := range allEntries {
				if i == removeIdx {
					// 削除されたエントリは存在しないはず
					for _, configEntry := range config.Entries {
						if configEntry.Type == entry.Type &&
							configEntry.Name == entry.Name &&
							configEntry.Definition == entry.Definition {
							t.Logf("Removed entry still exists: type=%s, name=%s",
								entry.Type, entry.Name)
							return false
						}
					}
				} else {
					// 他のエントリは存在するはず
					found := false
					for _, configEntry := range config.Entries {
						if configEntry.Type == entry.Type &&
							configEntry.Name == entry.Name &&
							configEntry.Definition == entry.Definition {
							found = true
							break
						}
					}
					if !found {
						t.Logf("Non-target entry not preserved: type=%s, name=%s, definition=%s",
							entry.Type, entry.Name, entry.Definition)
						return false
					}
				}
			}

			return true
		}

		// プロパティテストを実行
		config := &quick.Config{
			MaxCount: 50, // 50回のランダムテストを実行
		}

		if err := quick.Check(property, config); err != nil {
			t.Errorf("Property violated: %v", err)
		}
	})
}

// TestProperty_EntryRemovalAccuracy tests Property 4: エントリ削除の正確性
// **Validates: Requirements 3.5**
//
// Property: For any Management_File and target entry to remove, when that entry is removed,
// the entry will not exist when the file is reloaded.
func TestProperty_EntryRemovalAccuracy(t *testing.T) {
	// プロパティ関数: エントリを削除して再読み込みすると、そのエントリが存在しない
	property := func(entryCount uint8, removeIndex uint8, entryType string, name string, definition string) bool {
		// エントリの総数を制限（1-5個）
		count := int(entryCount%5) + 1
		// 削除するインデックスを範囲内に制限
		removeIdx := int(removeIndex) % count

		// 削除対象エントリの値を正規化
		entryType = normalizeEntryType(entryType)
		name = normalizeName(name)
		definition = normalizeDefinition(definition)

		// 空白のみの入力はスキップ
		if strings.TrimSpace(name) == "" || strings.TrimSpace(definition) == "" {
			return true
		}

		// テスト用の一時ディレクトリを作成
		tempDir, err := os.MkdirTemp("", "fish-config-test-*")
		if err != nil {
			t.Logf("Failed to create temp dir: %v", err)
			return false
		}
		defer os.RemoveAll(tempDir)

		// テスト用のConfigManagerを作成
		testFilePath := filepath.Join(tempDir, "fish-configurator.fish")
		manager := &DefaultConfigManager{
			filePath: testFilePath,
			parser:   NewParser(),
		}

		// エントリを追加
		allEntries := make([]Entry, count)
		for i := 0; i < count; i++ {
			var eType, eName, eDef string
			
			// 削除対象のインデックスには指定された値を使用
			if i == removeIdx {
				eType = entryType
				eName = name
				eDef = definition
			} else {
				// 他のエントリには固定値を使用
				eType = "alias"
				if i%2 == 1 {
					eType = "abbr"
				}
				eName = fmt.Sprintf("entry%d", i)
				eDef = fmt.Sprintf("echo entry%d", i)
			}

			allEntries[i] = Entry{
				Type:       eType,
				Name:       eName,
				Definition: eDef,
			}

			err := manager.AddEntry(eType, eName, eDef)
			if err != nil {
				t.Logf("Failed to add entry: %v", err)
				return false
			}
		}

		// 削除対象のエントリを取得
		toRemove := allEntries[removeIdx]

		// エントリを削除
		err = manager.RemoveEntry(toRemove.Type, toRemove.Name)
		if err != nil {
			t.Logf("Failed to remove entry: %v", err)
			return false
		}

		// ファイルを再読み込み
		config, err := manager.Load()
		if err != nil {
			t.Logf("Failed to reload config: %v", err)
			return false
		}

		// 削除したエントリが存在しないことを確認
		for _, entry := range config.Entries {
			if entry.Type == toRemove.Type && 
				entry.Name == toRemove.Name && 
				entry.Definition == toRemove.Definition {
				t.Logf("Removed entry still exists: type=%s, name=%s, definition=%s",
					toRemove.Type, toRemove.Name, toRemove.Definition)
				return false
			}
		}

		return true
	}

	// プロパティテストを実行
	config := &quick.Config{
		MaxCount: 100, // 100回のランダムテストを実行
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property violated: %v", err)
	}
}

// TestProperty_FormatAccuracy tests Property 7: フォーマット正確性
// **Validates: Requirements 5.5, 5.6**
//
// Property: For any entry (type, name, definition), when writing to Management_File,
// alias entries use the format "alias <name> '<definition>'" and
// abbr entries use the format "abbr -a <name> '<definition>'".
func TestProperty_FormatAccuracy(t *testing.T) {
	// プロパティ関数: エントリを追加すると、正しいフォーマットでファイルに書き込まれる
	property := func(entryType string, name string, definition string) bool {
		// 入力を正規化して有効な値にする
		entryType = normalizeEntryType(entryType)
		name = normalizeName(name)
		definition = normalizeDefinition(definition)

		// 空白のみの入力はスキップ
		if strings.TrimSpace(name) == "" || strings.TrimSpace(definition) == "" {
			return true
		}

		// テスト用の一時ディレクトリを作成
		tempDir, err := os.MkdirTemp("", "fish-config-test-*")
		if err != nil {
			t.Logf("Failed to create temp dir: %v", err)
			return false
		}
		defer os.RemoveAll(tempDir)

		// テスト用のConfigManagerを作成
		testFilePath := filepath.Join(tempDir, "fish-configurator.fish")
		manager := &DefaultConfigManager{
			filePath: testFilePath,
			parser:   NewParser(),
		}

		// エントリを追加
		err = manager.AddEntry(entryType, name, definition)
		if err != nil {
			t.Logf("Failed to add entry: %v", err)
			return false
		}

		// ファイルの内容を読み込む
		content, err := os.ReadFile(testFilePath)
		if err != nil {
			t.Logf("Failed to read file: %v", err)
			return false
		}

		fileContent := string(content)

		// 期待されるフォーマットを構築
		var expectedLine string
		if entryType == "alias" {
			expectedLine = fmt.Sprintf("alias %s '%s'", name, definition)
		} else if entryType == "abbr" {
			expectedLine = fmt.Sprintf("abbr -a %s '%s'", name, definition)
		} else {
			t.Logf("Unknown entry type: %s", entryType)
			return false
		}

		// ファイル内容に期待される行が含まれているか確認
		if !strings.Contains(fileContent, expectedLine) {
			t.Logf("Expected format not found in file.\nExpected line: %s\nFile content:\n%s",
				expectedLine, fileContent)
			return false
		}

		// パーサーで解析して、正しく読み込めることを確認
		config, err := manager.Load()
		if err != nil {
			t.Logf("Failed to parse file: %v", err)
			return false
		}

		// 追加したエントリが正しく解析されることを確認
		found := false
		for _, entry := range config.Entries {
			if entry.Type == entryType && entry.Name == name && entry.Definition == definition {
				found = true
				break
			}
		}

		if !found {
			t.Logf("Entry not found after parsing: type=%s, name=%s, definition=%s",
				entryType, name, definition)
			return false
		}

		return true
	}

	// プロパティテストを実行
	config := &quick.Config{
		MaxCount: 100, // 100回のランダムテストを実行
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property violated: %v", err)
	}
}

// TestProperty_FilePersistenceAccuracy tests Property 10: ファイル永続化の正確性
// **Validates: Requirements 9.1, 9.2**
//
// Property: For any entry addition or removal operation, when the operation succeeds,
// changes are immediately reflected in Management_File and file integrity is guaranteed.
func TestProperty_FilePersistenceAccuracy(t *testing.T) {
	t.Run("AddEntry immediately persists to file", func(t *testing.T) {
		// プロパティ関数: エントリを追加すると、即座にファイルに反映される
		property := func(entryType string, name string, definition string) bool {
			// 入力を正規化して有効な値にする
			entryType = normalizeEntryType(entryType)
			name = normalizeName(name)
			definition = normalizeDefinition(definition)

			// 空白のみの入力はスキップ
			if strings.TrimSpace(name) == "" || strings.TrimSpace(definition) == "" {
				return true
			}

			// テスト用の一時ディレクトリを作成
			tempDir, err := os.MkdirTemp("", "fish-config-test-*")
			if err != nil {
				t.Logf("Failed to create temp dir: %v", err)
				return false
			}
			defer os.RemoveAll(tempDir)

			// テスト用のConfigManagerを作成
			testFilePath := filepath.Join(tempDir, "fish-configurator.fish")
			manager := &DefaultConfigManager{
				filePath: testFilePath,
				parser:   NewParser(),
			}

			// エントリを追加
			err = manager.AddEntry(entryType, name, definition)
			if err != nil {
				t.Logf("Failed to add entry: %v", err)
				return false
			}

			// ファイルが存在することを確認（要件 9.1: 変更を即座に反映）
			if _, err := os.Stat(testFilePath); os.IsNotExist(err) {
				t.Logf("Management_File does not exist after AddEntry")
				return false
			}

			// ファイルが読み込み可能であることを確認（要件 9.2: ファイルの整合性を保証）
			content, err := os.ReadFile(testFilePath)
			if err != nil {
				t.Logf("Failed to read Management_File after AddEntry: %v", err)
				return false
			}

			// ファイル内容が空でないことを確認
			if len(content) == 0 {
				t.Logf("Management_File is empty after AddEntry")
				return false
			}

			// 新しいConfigManagerインスタンスを作成して、ファイルから読み込む
			// （これにより、メモリ上のキャッシュではなく、実際にファイルに書き込まれたことを確認）
			newManager := &DefaultConfigManager{
				filePath: testFilePath,
				parser:   NewParser(),
			}

			config, err := newManager.Load()
			if err != nil {
				t.Logf("Failed to load config from persisted file: %v", err)
				return false
			}

			// 追加したエントリが存在することを確認
			found := false
			for _, entry := range config.Entries {
				if entry.Type == entryType && entry.Name == name && entry.Definition == definition {
					found = true
					break
				}
			}

			if !found {
				t.Logf("Entry not found in persisted file: type=%s, name=%s, definition=%s",
					entryType, name, definition)
				return false
			}

			return true
		}

		// プロパティテストを実行
		config := &quick.Config{
			MaxCount: 100, // 100回のランダムテストを実行
		}

		if err := quick.Check(property, config); err != nil {
			t.Errorf("Property violated: %v", err)
		}
	})

	t.Run("RemoveEntry immediately persists to file", func(t *testing.T) {
		// プロパティ関数: エントリを削除すると、即座にファイルに反映される
		property := func(entryCount uint8, removeIndex uint8) bool {
			// エントリの総数を制限（1-5個）
			count := int(entryCount%5) + 1
			// 削除するインデックスを範囲内に制限
			removeIdx := int(removeIndex) % count

			// テスト用の一時ディレクトリを作成
			tempDir, err := os.MkdirTemp("", "fish-config-test-*")
			if err != nil {
				t.Logf("Failed to create temp dir: %v", err)
				return false
			}
			defer os.RemoveAll(tempDir)

			// テスト用のConfigManagerを作成
			testFilePath := filepath.Join(tempDir, "fish-configurator.fish")
			manager := &DefaultConfigManager{
				filePath: testFilePath,
				parser:   NewParser(),
			}

			// エントリを追加
			allEntries := make([]Entry, count)
			for i := 0; i < count; i++ {
				entryType := "alias"
				if i%2 == 1 {
					entryType = "abbr"
				}
				name := fmt.Sprintf("entry%d", i)
				definition := fmt.Sprintf("echo entry%d", i)

				allEntries[i] = Entry{
					Type:       entryType,
					Name:       name,
					Definition: definition,
				}

				err := manager.AddEntry(entryType, name, definition)
				if err != nil {
					t.Logf("Failed to add entry: %v", err)
					return false
				}
			}

			// 削除対象のエントリを取得
			toRemove := allEntries[removeIdx]

			// エントリを削除
			err = manager.RemoveEntry(toRemove.Type, toRemove.Name)
			if err != nil {
				t.Logf("Failed to remove entry: %v", err)
				return false
			}

			// ファイルが存在することを確認（要件 9.1: 変更を即座に反映）
			if _, err := os.Stat(testFilePath); os.IsNotExist(err) {
				t.Logf("Management_File does not exist after RemoveEntry")
				return false
			}

			// ファイルが読み込み可能であることを確認（要件 9.2: ファイルの整合性を保証）
			content, err := os.ReadFile(testFilePath)
			if err != nil {
				t.Logf("Failed to read Management_File after RemoveEntry: %v", err)
				return false
			}

			// ファイル内容が有効であることを確認（空でも良い場合がある）
			_ = content

			// 新しいConfigManagerインスタンスを作成して、ファイルから読み込む
			newManager := &DefaultConfigManager{
				filePath: testFilePath,
				parser:   NewParser(),
			}

			config, err := newManager.Load()
			if err != nil {
				t.Logf("Failed to load config from persisted file: %v", err)
				return false
			}

			// 削除したエントリが存在しないことを確認
			for _, entry := range config.Entries {
				if entry.Type == toRemove.Type &&
					entry.Name == toRemove.Name &&
					entry.Definition == toRemove.Definition {
					t.Logf("Removed entry still exists in persisted file: type=%s, name=%s",
						toRemove.Type, toRemove.Name)
					return false
				}
			}

			// 残りのエントリが正しく保持されていることを確認
			expectedCount := count - 1
			if len(config.Entries) != expectedCount {
				t.Logf("Expected %d entries after removal, got %d", expectedCount, len(config.Entries))
				return false
			}

			return true
		}

		// プロパティテストを実行
		config := &quick.Config{
			MaxCount: 100, // 100回のランダムテストを実行
		}

		if err := quick.Check(property, config); err != nil {
			t.Errorf("Property violated: %v", err)
		}
	})

	t.Run("File integrity is maintained across operations", func(t *testing.T) {
		// プロパティ関数: 複数の操作を行っても、ファイルの整合性が保たれる
		property := func(operationCount uint8) bool {
			// 操作回数を制限（1-10回）
			count := int(operationCount%10) + 1

			// テスト用の一時ディレクトリを作成
			tempDir, err := os.MkdirTemp("", "fish-config-test-*")
			if err != nil {
				t.Logf("Failed to create temp dir: %v", err)
				return false
			}
			defer os.RemoveAll(tempDir)

			// テスト用のConfigManagerを作成
			testFilePath := filepath.Join(tempDir, "fish-configurator.fish")
			manager := &DefaultConfigManager{
				filePath: testFilePath,
				parser:   NewParser(),
			}

			// 複数の操作を実行
			for i := 0; i < count; i++ {
				// 追加操作
				entryType := "alias"
				if i%2 == 1 {
					entryType = "abbr"
				}
				name := fmt.Sprintf("entry%d", i)
				definition := fmt.Sprintf("echo entry%d", i)

				err := manager.AddEntry(entryType, name, definition)
				if err != nil {
					t.Logf("Failed to add entry in operation %d: %v", i, err)
					return false
				}

				// 各操作後にファイルの整合性を確認
				if _, err := os.Stat(testFilePath); os.IsNotExist(err) {
					t.Logf("Management_File does not exist after operation %d", i)
					return false
				}

				// ファイルが読み込み可能であることを確認
				newManager := &DefaultConfigManager{
					filePath: testFilePath,
					parser:   NewParser(),
				}

				config, err := newManager.Load()
				if err != nil {
					t.Logf("Failed to load config after operation %d: %v", i, err)
					return false
				}

				// エントリ数が期待通りであることを確認
				expectedCount := i + 1
				if len(config.Entries) != expectedCount {
					t.Logf("Expected %d entries after operation %d, got %d",
						expectedCount, i, len(config.Entries))
					return false
				}
			}

			return true
		}

		// プロパティテストを実行
		config := &quick.Config{
			MaxCount: 50, // 50回のランダムテストを実行
		}

		if err := quick.Check(property, config); err != nil {
			t.Errorf("Property violated: %v", err)
		}
	})
}


// TestProperty_FileSaveVerification tests Property 11: ファイル保存の検証
// **Validates: Requirements 9.3**
//
// Property: For any write operation to Management_File, after the write completes,
// it can be confirmed that the file was saved correctly (file exists and is readable).
func TestProperty_FileSaveVerification(t *testing.T) {
	t.Run("AddEntry verifies file is saved correctly", func(t *testing.T) {
		// プロパティ関数: エントリを追加した後、ファイルが正しく保存されたことを確認できる
		property := func(entryType string, name string, definition string) bool {
			// 入力を正規化して有効な値にする
			entryType = normalizeEntryType(entryType)
			name = normalizeName(name)
			definition = normalizeDefinition(definition)

			// 空白のみの入力はスキップ
			if strings.TrimSpace(name) == "" || strings.TrimSpace(definition) == "" {
				return true
			}

			// テスト用の一時ディレクトリを作成
			tempDir, err := os.MkdirTemp("", "fish-config-test-*")
			if err != nil {
				t.Logf("Failed to create temp dir: %v", err)
				return false
			}
			defer os.RemoveAll(tempDir)

			// テスト用のConfigManagerを作成
			testFilePath := filepath.Join(tempDir, "fish-configurator.fish")
			manager := &DefaultConfigManager{
				filePath: testFilePath,
				parser:   NewParser(),
			}

			// エントリを追加
			err = manager.AddEntry(entryType, name, definition)
			if err != nil {
				t.Logf("Failed to add entry: %v", err)
				return false
			}

			// 要件 9.3: ファイルが正しく保存されたことを確認する
			// 1. ファイルが存在することを確認
			fileInfo, err := os.Stat(testFilePath)
			if err != nil {
				if os.IsNotExist(err) {
					t.Logf("File does not exist after write operation")
					return false
				}
				t.Logf("Failed to stat file: %v", err)
				return false
			}

			// 2. ファイルが通常のファイルであることを確認
			if !fileInfo.Mode().IsRegular() {
				t.Logf("File is not a regular file")
				return false
			}

			// 3. ファイルが読み込み可能であることを確認
			content, err := os.ReadFile(testFilePath)
			if err != nil {
				t.Logf("File exists but is not readable: %v", err)
				return false
			}

			// 4. ファイル内容が空でないことを確認
			if len(content) == 0 {
				t.Logf("File is empty after write operation")
				return false
			}

			return true
		}

		// プロパティテストを実行
		config := &quick.Config{
			MaxCount: 100, // 100回のランダムテストを実行
		}

		if err := quick.Check(property, config); err != nil {
			t.Errorf("Property violated: %v", err)
		}
	})

	t.Run("RemoveEntry verifies file is saved correctly", func(t *testing.T) {
		// プロパティ関数: エントリを削除した後、ファイルが正しく保存されたことを確認できる
		property := func(entryCount uint8, removeIndex uint8) bool {
			// エントリの総数を制限（1-5個）
			count := int(entryCount%5) + 1
			// 削除するインデックスを範囲内に制限
			removeIdx := int(removeIndex) % count

			// テスト用の一時ディレクトリを作成
			tempDir, err := os.MkdirTemp("", "fish-config-test-*")
			if err != nil {
				t.Logf("Failed to create temp dir: %v", err)
				return false
			}
			defer os.RemoveAll(tempDir)

			// テスト用のConfigManagerを作成
			testFilePath := filepath.Join(tempDir, "fish-configurator.fish")
			manager := &DefaultConfigManager{
				filePath: testFilePath,
				parser:   NewParser(),
			}

			// エントリを追加
			allEntries := make([]Entry, count)
			for i := 0; i < count; i++ {
				entryType := "alias"
				if i%2 == 1 {
					entryType = "abbr"
				}
				name := fmt.Sprintf("entry%d", i)
				definition := fmt.Sprintf("echo entry%d", i)

				allEntries[i] = Entry{
					Type:       entryType,
					Name:       name,
					Definition: definition,
				}

				err := manager.AddEntry(entryType, name, definition)
				if err != nil {
					t.Logf("Failed to add entry: %v", err)
					return false
				}
			}

			// 削除対象のエントリを取得
			toRemove := allEntries[removeIdx]

			// エントリを削除
			err = manager.RemoveEntry(toRemove.Type, toRemove.Name)
			if err != nil {
				t.Logf("Failed to remove entry: %v", err)
				return false
			}

			// 要件 9.3: ファイルが正しく保存されたことを確認する
			// 1. ファイルが存在することを確認
			fileInfo, err := os.Stat(testFilePath)
			if err != nil {
				if os.IsNotExist(err) {
					t.Logf("File does not exist after write operation")
					return false
				}
				t.Logf("Failed to stat file: %v", err)
				return false
			}

			// 2. ファイルが通常のファイルであることを確認
			if !fileInfo.Mode().IsRegular() {
				t.Logf("File is not a regular file")
				return false
			}

			// 3. ファイルが読み込み可能であることを確認
			content, err := os.ReadFile(testFilePath)
			if err != nil {
				t.Logf("File exists but is not readable: %v", err)
				return false
			}

			// 4. ファイル内容が有効であることを確認（空でも良い）
			_ = content

			return true
		}

		// プロパティテストを実行
		config := &quick.Config{
			MaxCount: 100, // 100回のランダムテストを実行
		}

		if err := quick.Check(property, config); err != nil {
			t.Errorf("Property violated: %v", err)
		}
	})

	t.Run("Save operation verifies file is saved correctly", func(t *testing.T) {
		// プロパティ関数: Save操作の後、ファイルが正しく保存されたことを確認できる
		property := func(entryCount uint8) bool {
			// エントリの総数を制限（0-10個）
			count := int(entryCount % 11)

			// テスト用の一時ディレクトリを作成
			tempDir, err := os.MkdirTemp("", "fish-config-test-*")
			if err != nil {
				t.Logf("Failed to create temp dir: %v", err)
				return false
			}
			defer os.RemoveAll(tempDir)

			// テスト用のConfigManagerを作成
			testFilePath := filepath.Join(tempDir, "fish-configurator.fish")
			manager := &DefaultConfigManager{
				filePath: testFilePath,
				parser:   NewParser(),
			}

			// Configを作成
			config := &Config{
				Entries: make([]Entry, count),
			}

			for i := 0; i < count; i++ {
				entryType := "alias"
				if i%2 == 1 {
					entryType = "abbr"
				}
				config.Entries[i] = Entry{
					Type:       entryType,
					Name:       fmt.Sprintf("entry%d", i),
					Definition: fmt.Sprintf("echo entry%d", i),
				}
			}

			// Configを保存
			err = manager.Save(config)
			if err != nil {
				t.Logf("Failed to save config: %v", err)
				return false
			}

			// 要件 9.3: ファイルが正しく保存されたことを確認する
			// 1. ファイルが存在することを確認
			fileInfo, err := os.Stat(testFilePath)
			if err != nil {
				if os.IsNotExist(err) {
					t.Logf("File does not exist after save operation")
					return false
				}
				t.Logf("Failed to stat file: %v", err)
				return false
			}

			// 2. ファイルが通常のファイルであることを確認
			if !fileInfo.Mode().IsRegular() {
				t.Logf("File is not a regular file")
				return false
			}

			// 3. ファイルが読み込み可能であることを確認
			content, err := os.ReadFile(testFilePath)
			if err != nil {
				t.Logf("File exists but is not readable: %v", err)
				return false
			}

			// 4. ファイル内容が有効であることを確認
			_ = content

			// 5. ファイルが解析可能であることを確認
			newManager := &DefaultConfigManager{
				filePath: testFilePath,
				parser:   NewParser(),
			}

			loadedConfig, err := newManager.Load()
			if err != nil {
				t.Logf("File is not parseable after save: %v", err)
				return false
			}

			// 6. 保存したエントリ数と読み込んだエントリ数が一致することを確認
			if len(loadedConfig.Entries) != count {
				t.Logf("Expected %d entries, got %d after save and load", count, len(loadedConfig.Entries))
				return false
			}

			return true
		}

		// プロパティテストを実行
		config := &quick.Config{
			MaxCount: 100, // 100回のランダムテストを実行
		}

		if err := quick.Check(property, config); err != nil {
			t.Errorf("Property violated: %v", err)
		}
	})
}
