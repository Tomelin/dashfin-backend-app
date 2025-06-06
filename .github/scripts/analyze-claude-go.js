const fs = require('fs');
const https = require('https');

async function analyzeGoWithClaude() {
  const changedFiles = process.argv[2].split(',').filter(f => f.trim());
  const prNumber = process.env.PR_NUMBER;
  const repoName = process.env.REPO_FULL_NAME;
  const claudeApiKey = process.env.CLAUDE_API_KEY;
  
  if (!claudeApiKey) {
    console.log('CLAUDE_API_KEY not found, skipping Claude review');
    return;
  }
  
  console.log(`Analyzing ${changedFiles.length} changed Go files with Claude...`);
  
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
  
  const prompt = `As an expert Go code reviewer, please analyze the following Go code changes with focus on:

## 🐹 Go Language & Idioms:
1. **Idiomatic Go**: Following Go conventions, effective Go patterns, and community best practices
2. **Error Handling**: Proper error wrapping, custom error types, and error propagation patterns
3. **Concurrency**: Goroutines, channels, select statements, and race condition prevention
4. **Interfaces**: Interface design, implicit implementation, and composition over inheritance
5. **Package Design**: Package organization, exported vs unexported identifiers, and API design

## 🏗️ Code Structure & Architecture:
1. **Function Design**: Single responsibility, appropriate function size, and clear parameter/return patterns
2. **Struct Composition**: Embedding, method sets, and value vs pointer receivers
3. **Context Usage**: Proper context.Context usage for cancellation and request-scoped values
4. **Dependency Injection**: Constructor patterns and dependency management

## ⚡ Performance & Memory:
1. **Memory Efficiency**: Slice capacity management, string builder usage, and memory pooling
2. **Algorithm Efficiency**: Big O complexity and optimization opportunities
3. **Garbage Collection**: Reducing allocations and GC pressure
4. **Profiling Considerations**: CPU and memory profiling opportunities

## 🛡️ Security & Reliability:
1. **Input Validation**: Parameter validation, boundary checks, and sanitization
2. **Resource Management**: Proper cleanup with defer, file/connection handling
3. **Cryptography**: Secure random generation, proper hashing, and encryption usage
4. **SQL Security**: Query parameterization and injection prevention

## 🧪 Testing & Quality:
1. **Test Patterns**: Table-driven tests, test helpers, and test organization
2. **Coverage**: Test completeness and edge case handling
3. **Benchmarks**: Performance testing opportunities
4. **Documentation**: GoDoc comments and code clarity

## 📦 Dependencies & Modules:
1. **Go Modules**: Proper versioning, dependency management, and module organization
2. **Standard Library**: Effective use of stdlib packages vs external dependencies
3. **Third-party Libraries**: Appropriate library choices and version management

Please provide specific, actionable feedback with Go code examples where helpful.

Code to review (${fileCount} files):
${codeContent.substring(0, 22000)}`; // Increased limit for Go files
  
  try {
    const response = await callClaudeAPI(prompt, claudeApiKey);
    
    if (response && response.trim()) {
      console.log('Claude Go analysis completed successfully');
      console.log('Response preview:', response.substring(0, 200) + '...');
      
      await postComment(prNumber, repoName, `## 🧠 Claude Go Code Review

${response}

---
*Analyzed ${fileCount} Go file(s) automatically*`);
    } else {
      console.log('Claude returned empty response');
    }
  } catch (error) {
    console.error('Claude analysis failed:', error.message);
    console.error('Error details:', error);
    
    await postComment(prNumber, repoName, `## 🧠 Claude Go Code Review

⚠️ **Analysis Failed**: Unable to complete Claude Go code review due to an API error.

Please check the workflow logs for more details.

---
*Attempted to analyze ${fileCount} Go file(s)*`);
  }
}

function callClaudeAPI(prompt, apiKey) {
  return new Promise((resolve, reject) => {
    const data = JSON.stringify({
      model: "claude-sonnet-4-20250514",
      max_tokens: 2048,
      temperature: 0.1,
      messages: [{
        role: "user",
        content: prompt
      }]
    });
    
    const options = {
      hostname: 'api.anthropic.com',
      path: '/v1/messages',
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'x-api-key': apiKey,
        'anthropic-version': '2023-06-01',
        'Content-Length': Buffer.byteLength(data)
      }
    };
    
    const req = https.request(options, (res) => {
      let body = '';
      res.on('data', chunk => body += chunk);
      res.on('end', () => {
        try {
          const result = JSON.parse(body);
          if (result.content && result.content[0]) {
            resolve(result.content[0].text);
          } else if (result.error) {
            reject(new Error(`Claude API error: ${result.error.message}`));
          } else {
            reject(new Error('Invalid Claude API response'));
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

analyzeGoWithClaude().catch(console.error);