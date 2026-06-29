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
});
