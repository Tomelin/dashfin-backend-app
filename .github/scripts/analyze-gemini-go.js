const fs = require('fs');
const https = require('https');

async function analyzeGoWithGemini() {
  const changedFiles = process.argv[2].split(',').filter(f => f.trim());
  const prNumber = process.env.PR_NUMBER;
  const repoName = process.env.REPO_FULL_NAME;
  const geminiApiKey = process.env.GEMINI_API_KEY;
  
  if (!geminiApiKey) {
    console.log('GEMINI_API_KEY not found, skipping Gemini review');
    return;
  }
  
  console.log(`Analyzing ${changedFiles.length} changed Go files with Gemini...`);
  
  // Read file contents
  let codeContent = '';
  let fileCount = 0;
  
  for (const file of changedFiles.slice(0, 15)) { // Limit to 15 files for Go
    try {
      if (fs.existsSync(file)) {
        const content = fs.readFileSync(file, 'utf8');
        codeContent += `\n\n--- File: ${file} ---\n${content}`;
        fileCount++;
      }
    } catch (err) {
      console.log(`Error reading ${file}:`, err.message);
    }
  }
  
  if (!codeContent.trim()) {
    console.log('No readable files found for analysis');
    return;
  }
  
  const prompt = `Please review the following Go code changes focusing on:

**🐹 Go Language Specifics:**
- Idiomatic Go patterns and conventions
- Proper error handling with explicit error returns
- Goroutine usage and concurrent programming patterns
- Channel usage and select statements
- Interface design and implementation
- Struct composition vs inheritance patterns
- Package organization and naming conventions

**📦 Go Modules & Dependencies:**
- go.mod and go.sum file changes
- Dependency management best practices
- Version compatibility issues
- Unused dependencies

**🔧 Code Quality:**
- Function and variable naming (camelCase, exported vs unexported)
- Code organization and structure
- Comment quality (especially for exported functions)
- Test coverage and test patterns
- Benchmark tests where applicable

**⚡ Performance & Memory:**
- Memory allocation patterns
- Slice and map usage efficiency
- String manipulation optimization
- Context usage for cancellation and timeouts
- Resource cleanup (defer statements)

**🛡️ Security & Best Practices:**
- Input validation and sanitization
- SQL injection prevention
- Cross-site scripting (XSS) prevention
- Proper use of crypto packages
- File path traversal prevention

**🚀 Standard Library Usage:**
- Effective use of standard library packages
- HTTP handler patterns and middleware
- JSON marshaling/unmarshaling
- Time and date handling
- Regular expressions

**🧪 Testing:**
- Table-driven tests
- Test helper functions
- Mocking and dependency injection
- Integration test patterns
- Error case testing

Files analyzed (${fileCount}):
${codeContent.substring(0, 18000)}`; // Increased limit for Go files
  
  try {
    const response = await callGeminiAPI(prompt, geminiApiKey);
    
    if (response && response.trim()) {
      console.log('Gemini Go analysis completed successfully');
      console.log('Response preview:', response.substring(0, 200) + '...');
      
      await postComment(prNumber, repoName, `## 🐹 Gemini Go Code Review

${response}

---
*Analyzed ${fileCount} Go file(s) automatically*`);
    } else {
      console.log('Gemini returned empty response');
    }
  } catch (error) {
    console.error('Gemini analysis failed:', error.message);
    console.error('Error details:', error);
    
    await postComment(prNumber, repoName, `## 🐹 Gemini Go Code Review

⚠️ **Analysis Failed**: Unable to complete Gemini Go code review due to an API error.

Please check the workflow logs for more details.

---
*Attempted to analyze ${fileCount} Go file(s)*`);
  }
}

function callGeminiAPI(prompt, apiKey) {
  return new Promise((resolve, reject) => {
    const data = JSON.stringify({
      contents: [{
        parts: [{ text: prompt }]
      }],
      generationConfig: {
        temperature: 0.1,
        maxOutputTokens: 2048
      }
    });
    
    const options = {
      hostname: 'generativelanguage.googleapis.com',
      path: `/v1beta/models/gemini-1.5-flash-latest:generateContent?key=${apiKey}`,
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Content-Length': Buffer.byteLength(data)
      }
    };
    
    const req = https.request(options, (res) => {
      let body = '';
      res.on('data', chunk => body += chunk);
      res.on('end', () => {
        try {
          const result = JSON.parse(body);
          if (result.candidates && result.candidates[0]) {
            resolve(result.candidates[0].content.parts[0].text);
          } else {
            reject(new Error('Invalid Gemini API response'));
          }
        } catch (e) {
          reject(new Error(`Parse error: ${e.message}`));
        }
      });
    });
    
    req.on('error', reject);
    req.write(data);
    req.end();
  });
}

async function postComment(prNumber, repoName, comment) {
  return new Promise((resolve, reject) => {
    if (!prNumber || !repoName || !process.env.GITHUB_TOKEN) {
      console.log('Missing required parameters for posting comment');
      console.log(`PR: ${prNumber}, Repo: ${repoName}, Token: ${process.env.GITHUB_TOKEN ? 'present' : 'missing'}`);
      resolve();
      return;
    }
    
    const data = JSON.stringify({ body: comment });
    
    const options = {
      hostname: 'api.github.com',
      path: `/repos/${repoName}/issues/${prNumber}/comments`,
      method: 'POST',
      headers: {
        'Authorization': `token ${process.env.GITHUB_TOKEN}`,
        'Accept': 'application/vnd.github.v3+json',
        'User-Agent': 'Go-AI-Code-Review-Action',
        'Content-Type': 'application/json',
        'Content-Length': Buffer.byteLength(data)
      }
    };
    
    console.log(`Posting comment to: ${options.hostname}${options.path}`);
    
    const req = https.request(options, (res) => {
      let body = '';
      res.on('data', chunk => body += chunk);
      res.on('end', () => {
        if (res.statusCode >= 200 && res.statusCode < 300) {
          console.log('Comment posted successfully');
          resolve();
        } else {
          console.error(`GitHub API error: ${res.statusCode} - ${body}`);
          resolve();
        }
      });
    });
    
    req.on('error', (error) => {
      console.error('Request error:', error.message);
      resolve();
    });
    
    req.write(data);
    req.end();
  });
}

analyzeGoWithGemini().catch(console.error);