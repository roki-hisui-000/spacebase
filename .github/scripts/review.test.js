const test = require('node:test');
const assert = require('node:assert');
const fs = require('fs');
const { run } = require('./review');

test.describe('review.js', () => {
  const originalEnv = { ...process.env };

  test.beforeEach(() => {
    process.env = { ...originalEnv };
    process.env.GEMINI_API_KEY = 'test-api-key';
  });

  test.it('GEMINI_API_KEYが設定されていない場合、エラーを出力して終了する', async (t) => {
    delete process.env.GEMINI_API_KEY;

    const mockConsoleError = t.mock.method(console, 'error', () => {});
    const mockExit = t.mock.method(process, 'exit', () => {});

    await run();

    assert.strictEqual(mockConsoleError.mock.calls.length, 1);
    assert.strictEqual(mockConsoleError.mock.calls[0].arguments[0], "GEMINI_API_KEY is not set");
    assert.strictEqual(mockExit.mock.calls.length, 1);
    assert.strictEqual(mockExit.mock.calls[0].arguments[0], 1);
  });

  test.it('diff.txt が存在しない場合、処理をスキップする', async (t) => {
    const mockExistsSync = t.mock.method(fs, 'existsSync', (path) => {
      if (path === 'diff.txt') return false;
      return true;
    });
    const mockConsoleLog = t.mock.method(console, 'log', () => {});
    const mockExit = t.mock.method(process, 'exit', () => {});

    await run();

    assert.strictEqual(mockConsoleLog.mock.calls.length, 1);
    assert.strictEqual(mockConsoleLog.mock.calls[0].arguments[0], "No diff found");
    assert.strictEqual(mockExit.mock.calls.length, 0); // スキップなので正常終了（exitは呼ばれない）
  });

  test.it('設定ファイルやドキュメントのみの変更の場合、処理をスキップする', async (t) => {
    const mockExistsSync = t.mock.method(fs, 'existsSync', () => true);
    const mockReadFileSync = t.mock.method(fs, 'readFileSync', (path) => {
      if (path === 'diff.txt') {
        return `diff --git a/config/config.dev.yaml b/config/config.dev.yaml
index 1234567..abcdefg 100644
--- a/config/config.dev.yaml
+++ b/config/config.dev.yaml
@@ -1,3 +1,4 @@
+new_config: true
diff --git a/docs/rules/pr_review.md b/docs/rules/pr_review.md
index abcdefg..1234567 100644
--- a/docs/rules/pr_review.md
+++ b/docs/rules/pr_review.md
@@ -1,3 +1,4 @@
+new_rule: true`;
      }
      return '';
    });
    const mockConsoleLog = t.mock.method(console, 'log', () => {});
    const mockExit = t.mock.method(process, 'exit', () => {});

    const mockFetch = t.mock.method(global, 'fetch', () => {
      throw new Error('fetch should not be called');
    });

    await run();

    assert.strictEqual(mockConsoleLog.mock.calls.some(c => c.arguments[0] === "Review skipped: Config or document changes only"), true);
    assert.strictEqual(mockExit.mock.calls.length, 0);
  });

  test.it('設定ファイルやドキュメント以外の変更が含まれる場合、処理をスキップしない', async (t) => {
    const mockExistsSync = t.mock.method(fs, 'existsSync', () => true);
    const mockReadFileSync = t.mock.method(fs, 'readFileSync', (path) => {
      if (path === 'diff.txt') {
        return `diff --git a/config/config.dev.yaml b/config/config.dev.yaml
index 1234567..abcdefg 100644
--- a/config/config.dev.yaml
+++ b/config/config.dev.yaml
@@ -1,3 +1,4 @@
+new_config: true
diff --git a/main.go b/main.go
index abcdefg..1234567 100644
--- a/main.go
+++ b/main.go
@@ -1,3 +1,4 @@
+new_logic: true`;
      }
      return '';
    });
    const mockReaddirSync = t.mock.method(fs, 'readdirSync', () => []);
    const mockWriteFileSync = t.mock.method(fs, 'writeFileSync', () => {});
    const mockConsoleLog = t.mock.method(console, 'log', () => {});

    let fetchCallCount = 0;
    const mockFetch = t.mock.method(global, 'fetch', async (url) => {
      fetchCallCount++;
      if (fetchCallCount === 1) {
        return {
          ok: true,
          json: async () => ({
            models: [{ name: 'models/gemini-1.5-flash', supportedGenerationMethods: ['generateContent'] }]
          })
        };
      } else {
        return {
          ok: true,
          json: async () => ({
            candidates: [{ content: { parts: [{ text: 'Mocked Gemini Review Comment' }] } }]
          })
        };
      }
    });

    await run();

    assert.strictEqual(mockWriteFileSync.mock.calls.length, 1);
    assert.strictEqual(mockConsoleLog.mock.calls.some(c => c.arguments[0] === "Review generated successfully"), true);
  });

  test.it('docs/spec などの設計書ドキュメント変更が含まれる場合、処理をスキップしない', async (t) => {
    const mockExistsSync = t.mock.method(fs, 'existsSync', () => true);
    const mockReadFileSync = t.mock.method(fs, 'readFileSync', (path) => {
      if (path === 'diff.txt') {
        return `diff --git a/docs/spec/feature.md b/docs/spec/feature.md
index abcdefg..1234567 100644
--- a/docs/spec/feature.md
+++ b/docs/spec/feature.md
@@ -1,3 +1,4 @@
+新機能の設計仕様書`;
      }
      return '';
    });
    const mockReaddirSync = t.mock.method(fs, 'readdirSync', () => []);
    const mockWriteFileSync = t.mock.method(fs, 'writeFileSync', () => {});
    const mockConsoleLog = t.mock.method(console, 'log', () => {});

    let fetchCallCount = 0;
    const mockFetch = t.mock.method(global, 'fetch', async (url) => {
      fetchCallCount++;
      if (fetchCallCount === 1) {
        return {
          ok: true,
          json: async () => ({
            models: [{ name: 'models/gemini-1.5-flash', supportedGenerationMethods: ['generateContent'] }]
          })
        };
      } else {
        return {
          ok: true,
          json: async () => ({
            candidates: [{ content: { parts: [{ text: 'Mocked Gemini Review Comment for Spec' }] } }]
          })
        };
      }
    });

    await run();

    assert.strictEqual(mockWriteFileSync.mock.calls.length, 1);
    assert.strictEqual(mockConsoleLog.mock.calls.some(c => c.arguments[0] === "Review generated successfully"), true);
  });

  test.it('正常にモデルを自動選択してレビューを生成・ファイルに書き込む', async (t) => {
    const mockExistsSync = t.mock.method(fs, 'existsSync', () => true);
    const mockReadFileSync = t.mock.method(fs, 'readFileSync', (path) => {
      if (path === 'diff.txt') return 'dummy diff';
      if (path === '.clinerules') return 'dummy rules';
      return '';
    });
    const mockReaddirSync = t.mock.method(fs, 'readdirSync', () => []);
    const mockWriteFileSync = t.mock.method(fs, 'writeFileSync', () => {});
    const mockConsoleLog = t.mock.method(console, 'log', () => {});

    // fetch のモック化
    let fetchCallCount = 0;
    const mockFetch = t.mock.method(global, 'fetch', async (url) => {
      fetchCallCount++;
      if (fetchCallCount === 1) {
        // models API のレスポンス
        return {
          ok: true,
          json: async () => ({
            models: [
              { name: 'models/gemini-1.5-flash', supportedGenerationMethods: ['generateContent'] }
            ]
          })
        };
      } else {
        // generateContent API のレスポンス
        return {
          ok: true,
          json: async () => ({
            candidates: [{
              content: {
                parts: [{ text: 'Mocked Gemini Review Comment' }]
              }
            }]
          })
        };
      }
    });

    await run();

    assert.strictEqual(mockWriteFileSync.mock.calls.length, 1);
    assert.strictEqual(mockWriteFileSync.mock.calls[0].arguments[0], 'review_result.txt');
    assert.strictEqual(mockWriteFileSync.mock.calls[0].arguments[1], 'Mocked Gemini Review Comment');
    assert.strictEqual(mockConsoleLog.mock.calls.some(c => c.arguments[0] === "Review generated successfully"), true);
  });

  test.it('models API呼び出しが失敗した場合、警告ログを出力しデフォルトモデルを使用する', async (t) => {
    const mockExistsSync = t.mock.method(fs, 'existsSync', () => true);
    const mockReadFileSync = t.mock.method(fs, 'readFileSync', () => 'dummy content');
    const mockReaddirSync = t.mock.method(fs, 'readdirSync', () => []);
    const mockWriteFileSync = t.mock.method(fs, 'writeFileSync', () => {});
    const mockConsoleWarn = t.mock.method(console, 'warn', () => {});
    const mockConsoleLog = t.mock.method(console, 'log', () => {});

    let fetchCallCount = 0;
    const mockFetch = t.mock.method(global, 'fetch', async (url) => {
      fetchCallCount++;
      if (fetchCallCount === 1) {
        // models API のレスポンスを失敗させる
        return { ok: false, status: 500, text: async () => 'Internal Server Error' };
      } else {
        // generateContent API のレスポンス（デフォルトモデルが使われることを想定）
        assert.ok(url.includes('gemini-1.5-flash'), 'Should use default model if models API fails');
        return {
          ok: true,
          json: async () => ({
            candidates: [{ content: { parts: [{ text: 'Mocked Gemini Review Comment with default model' }] } }]
          })
        };
      }
    });

    await run();

    assert.strictEqual(mockConsoleWarn.mock.calls.length, 1);
    assert.match(mockConsoleWarn.mock.calls[0].arguments[0], /Failed to list models/);
    assert.strictEqual(mockWriteFileSync.mock.calls[0].arguments[1], 'Mocked Gemini Review Comment with default model');
    assert.strictEqual(mockConsoleLog.mock.calls.some(c => c.arguments[0].includes("Sending review request to Gemini model: gemini-1.5-flash")), true);
  });

  test.it('generateContent API呼び出しが失敗した場合、エラーを出力して終了する', async (t) => {
    const mockExistsSync = t.mock.method(fs, 'existsSync', () => true);
    const mockReadFileSync = t.mock.method(fs, 'readFileSync', () => 'dummy content');
    const mockReaddirSync = t.mock.method(fs, 'readdirSync', () => []);
    const mockConsoleError = t.mock.method(console, 'error', () => {});
    const mockExit = t.mock.method(process, 'exit', () => {});
    t.mock.method(console, 'log', () => {}); // 不要なログ出力を抑制

    let fetchCallCount = 0;
    const mockFetch = t.mock.method(global, 'fetch', async (url) => {
      fetchCallCount++;
      if (fetchCallCount === 1) {
        // models API のレスポンス (成功)
        return {
          ok: true,
          json: async () => ({ models: [{ name: 'models/gemini-1.5-flash', supportedGenerationMethods: ['generateContent'] }] })
        };
      } else {
        // generateContent API のレスポンスを失敗させる
        return { ok: false, status: 400, text: async () => 'Bad Request' };
      }
    });

    await run();

    assert.strictEqual(mockConsoleError.mock.calls.length, 1);
    assert.match(mockConsoleError.mock.calls[0].arguments[0], /Gemini API error/);
    assert.strictEqual(mockExit.mock.calls.length, 1);
    assert.strictEqual(mockExit.mock.calls[0].arguments[0], 1);
  });

  test.it('generateContent APIのレスポンス形式が不正な場合、エラーを出力して終了する', async (t) => {
    const mockExistsSync = t.mock.method(fs, 'existsSync', () => true);
    const mockReadFileSync = t.mock.method(fs, 'readFileSync', () => 'dummy content');
    const mockReaddirSync = t.mock.method(fs, 'readdirSync', () => []);
    const mockConsoleError = t.mock.method(console, 'error', () => {});
    const mockExit = t.mock.method(process, 'exit', () => {});
    t.mock.method(console, 'log', () => {}); // 不要なログ出力を抑制

    let fetchCallCount = 0;
    const mockFetch = t.mock.method(global, 'fetch', async (url) => {
      fetchCallCount++;
      if (fetchCallCount === 1) {
        // models API のレスポンス (成功)
        return {
          ok: true,
          json: async () => ({ models: [{ name: 'models/gemini-1.5-flash', supportedGenerationMethods: ['generateContent'] }] })
        };
      } else {
        // generateContent API のレスポンスが不正 (candidatesがない)
        return { ok: true, json: async () => ({ /* 空オブジェクトまたは不正な構造 */ }) };
      }
    });

    await run();

    assert.strictEqual(mockConsoleError.mock.calls.length, 1);
    assert.match(mockConsoleError.mock.calls[0].arguments[0], /Invalid response from Gemini API/);
    assert.strictEqual(mockExit.mock.calls.length, 1);
    assert.strictEqual(mockExit.mock.calls[0].arguments[0], 1);
  });

  test.it('docs/rules ディレクトリが存在しない場合、docRules が空になる', async (t) => {
    const mockExistsSync = t.mock.method(fs, 'existsSync', (path) => {
      if (path === 'diff.txt') return true;
      if (path === '.clinerules') return true;
      if (path === 'docs/rules') return false; // docs/rulesが存在しない
      return true;
    });
    const mockReadFileSync = t.mock.method(fs, 'readFileSync', (path) => {
      if (path === 'diff.txt') return 'dummy diff';
      if (path === '.clinerules') return 'dummy rules';
      return '';
    });
    
    // fetch のモック
    const mockFetch = t.mock.method(global, 'fetch', async (url, options) => {
      if (url.includes('/models?')) {
        return {
          ok: true,
          json: async () => ({ models: [{ name: 'models/gemini-1.5-flash', supportedGenerationMethods: ['generateContent'] }] })
        };
      } else {
        const body = JSON.parse(options.body);
        const prompt = body.contents[0].parts[0].text;
        // プロンプトにdocs/rulesに関する内容が含まれていないことをアサート
        assert.ok(!prompt.includes('--- architecture.md ---'), 'Prompt should not contain docRules content');
        return {
          ok: true,
          json: async () => ({
            candidates: [{ content: { parts: [{ text: 'Mocked Gemini Review Comment with no docRules' }] } }]
          })
        };
      }
    });

    const mockWriteFileSync = t.mock.method(fs, 'writeFileSync', () => {});
    const mockConsoleLog = t.mock.method(console, 'log', () => {});

    await run();

    assert.strictEqual(mockExistsSync.mock.calls.some(c => c.arguments[0] === 'docs/rules'), true);
    assert.strictEqual(mockWriteFileSync.mock.calls.length, 1);
    assert.strictEqual(mockWriteFileSync.mock.calls[0].arguments[1], 'Mocked Gemini Review Comment with no docRules');
    assert.strictEqual(mockConsoleLog.mock.calls.some(c => c.arguments[0] === "Review generated successfully"), true);
  });
});
