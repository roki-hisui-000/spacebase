const fs = require('fs');

async function run() {
  const apiKey = process.env.GEMINI_API_KEY;
  if (!apiKey) {
    console.error("GEMINI_API_KEY is not set");
    process.exit(1);
  }

  // 差分ファイルの読み込み
  const diff = fs.existsSync('diff.txt') ? fs.readFileSync('diff.txt', 'utf8') : '';
  if (!diff) {
    console.log("No diff found");
    return;
  }

  // ルールファイルの読み込み
  const clinerules = fs.existsSync('.clinerules') ? fs.readFileSync('.clinerules', 'utf8') : '';
  
  // docs/rules/*.md の読み込み
  let docRules = '';
  const rulesDir = 'docs/rules';
  if (fs.existsSync(rulesDir)) {
    const files = fs.readdirSync(rulesDir);
    for (const file of files) {
      if (file.endsWith('.md')) {
        docRules += `\n--- ${file} ---\n` + fs.readFileSync(`${rulesDir}/${file}`, 'utf8');
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

  const url = `https://generativelanguage.googleapis.com/v1beta/models/gemini-1.5-flash:generateContent?key=${apiKey}`;
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
    console.error(`Gemini API error: ${response.status}`, errText);
    process.exit(1);
  }

  const data = await response.json();
  const reviewResult = data.candidates?.[0]?.content?.parts?.[0]?.text;
  if (!reviewResult) {
    console.error("Invalid response from Gemini API", JSON.stringify(data));
    process.exit(1);
  }

  fs.writeFileSync('review_result.txt', reviewResult);
  console.log("Review generated successfully");
}

run().catch(err => {
  console.error(err);
  process.exit(1);
});
