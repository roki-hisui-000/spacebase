const fs = require('fs');

// 定数定義
const GEMINI_API_BASE_URL = 'https://generativelanguage.googleapis.com/v1beta';
const DEFAULT_GEMINI_MODEL = 'gemini-1.5-flash';

// ファイルパスおよびディレクトリ名の定数化
const DIFF_FILE_NAME = 'diff.txt';
const CLINE_RULES_FILE_NAME = '.clinerules';
const DOC_RULES_DIR = 'docs/rules';
const REVIEW_RESULT_FILE_NAME = 'review_result.txt';

// プログラムの終了コードの定数化
const EXIT_CODE_ERROR = 1;

async function run() {
  const apiKey = process.env.GEMINI_API_KEY;
  if (!apiKey) {
    console.error("GEMINI_API_KEY is not set");
    process.exit(EXIT_CODE_ERROR);
    return;
  }

  // 差分ファイルの読み込み
  const diff = fs.existsSync(DIFF_FILE_NAME) ? fs.readFileSync(DIFF_FILE_NAME, 'utf8') : '';
  if (!diff) {
    console.log("No diff found");
    return;
  }

  // ルールファイルの読み込み
  const clinerules = fs.existsSync(CLINE_RULES_FILE_NAME) ? fs.readFileSync(CLINE_RULES_FILE_NAME, 'utf8') : '';
  
  // docs/rules/*.md の読み込み
  let docRules = '';
  if (fs.existsSync(DOC_RULES_DIR)) {
    const files = fs.readdirSync(DOC_RULES_DIR);
    for (const file of files) {
      if (file.endsWith('.md')) {
        docRules += `\n--- ${file} ---\n` + fs.readFileSync(`${DOC_RULES_DIR}/${file}`, 'utf8');
      }
    }
  }

  const prompt = `あなたは当プロジェクトの極めて優秀なシニアエンジニア兼レビュアーです。
提出されたPRのコード差分（diff）を、以下の「開発ルール」に基づいて厳格にレビューしてください。

もしルールへの違反や、改善の余地がある場合は、該当する箇所へ具体的な修正案コードを添えて指摘を行ってください。

--- 【1】全体ルール (.clinerules) ---
${clinerules}

--- 【2】詳細・専門ルール ---
${docRules}

--- 【レビュー対象のコード差分 (diff)】 ---
${diff}
`;

  // 利用可能なモデルの一覧を取得し、自動的に最適なモデルを判定する（堅牢性の担保）
  let modelName = DEFAULT_GEMINI_MODEL; // デフォルトフォールバック
  try {
    const modelsUrl = `${GEMINI_API_BASE_URL}/models?key=${apiKey}`;
    const modelsRes = await fetch(modelsUrl);
    if (modelsRes.ok) {
      const modelsData = await modelsRes.json();
      const models = modelsData.models || [];
      console.log("Detected available models:", models.map(m => m.name));

      // 'generateContent' をサポートする flash モデルを探す（2.5や1.5等、最新順にマッチしやすいようフィルタ）
      const bestModel = models.find(m => 
        m.name.includes('gemini') && 
        m.name.includes('flash') && 
        m.supportedGenerationMethods?.includes('generateContent')
      );

      if (bestModel) {
        modelName = bestModel.name.replace('models/', '');
        console.log(`Auto-selected best model: ${modelName}`);
      } else {
        // flashが見つからない場合は、generateContentをサポートする任意のgeminiモデル
        const fallbackModel = models.find(m => 
          m.name.includes('gemini') && 
          m.supportedGenerationMethods?.includes('generateContent')
        );
        if (fallbackModel) {
          modelName = fallbackModel.name.replace('models/', '');
          console.log(`Auto-selected fallback model: ${modelName}`);
        }
      }
    } else {
      console.warn(`Failed to list models (status ${modelsRes.status}), using default: ${modelName}`);
    }
  } catch (err) {
    console.warn("Error while auto-detecting models, using default:", err);
  }

  const url = `${GEMINI_API_BASE_URL}/models/${modelName}:generateContent?key=${apiKey}`;
  console.log(`Sending review request to Gemini model: ${modelName}`);
  const response = await fetch(url, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      contents: [{
        parts: [{
          text: prompt
        }]
      }]
    })
  });

  if (!response.ok) {
    const errText = await response.text();
    console.error(`Gemini API error for model ${modelName}: ${response.status}`, errText);
    process.exit(EXIT_CODE_ERROR);
    return;
  }

  const data = await response.json();
  const reviewResult = data.candidates?.[0]?.content?.parts?.[0]?.text;
  if (!reviewResult) {
    console.error("Invalid response from Gemini API", JSON.stringify(data));
    process.exit(EXIT_CODE_ERROR);
    return;
  }

  fs.writeFileSync(REVIEW_RESULT_FILE_NAME, reviewResult);
  console.log("Review generated successfully");
}

if (require.main === module) {
  run().catch(err => {
    console.error(err);
    process.exit(1);
  });
}

module.exports = { run };
