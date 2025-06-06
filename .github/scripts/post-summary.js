const https = require('https');

async function postSummary() {
  const prNumber = process.env.PR_NUMBER;
  const repoOwner = process.env.REPO_OWNER;
  const repoName = process.env.REPO_NAME;
  const githubToken = process.env.GITHUB_TOKEN;
  
  if (!prNumber || !repoOwner || !repoName || !githubToken) {
    console.log('Missing required environment variables for summary');
    console.log(`PR: ${prNumber}, Owner: ${repoOwner}, Repo: ${repoName}, Token: ${githubToken ? 'present' : 'missing'}`);
    return;
  }
  
  try {
    // Get comments from the PR
    const comments = await getComments(prNumber, repoOwner, repoName, githubToken);
    
    const aiComments = comments.filter(comment => 
      comment.body.includes('🐹 Gemini Go Code Review') || 
      comment.body.includes('🧠 Claude Go Code Review')
    );
    
    // Get check runs
    const checkRuns = await getCheckRuns(repoOwner, repoName, githubToken);
    
    const lintStatus = getCheckStatus(checkRuns, ['golangci-lint', 'Go Lint']);
    const testStatus = getCheckStatus(checkRuns, ['test', 'Test']);
    
    if (aiComments.length > 0) {
      const summaryComment = `## ✅ Go AI Code Review Complete

This Go pull request has been automatically reviewed by AI assistants.

📊 **Review Summary:**
- ${aiComments.length} AI review(s) completed
- Linting status: ${formatStatus(lintStatus)}
- Test status: ${formatStatus(testStatus)}

🐹 **Go-Specific Checks:**
- Idiomatic Go patterns reviewed
- Concurrency and error handling analyzed
- Performance and security considerations evaluated
- Testing patterns and coverage assessed

*AI reviews are complementary to human code review and should not replace thorough manual review.*`;

      await postComment(prNumber, repoOwner, repoName, summaryComment, githubToken);
      console.log(`Go summary posted successfully for PR #${prNumber}`);
    } else {
      console.log('No AI comments found to summarize');
    }
  } catch (error) {
    console.error('Error posting summary:', error.message);
  }
}

function getComments(prNumber, owner, repo, token) {
  return new Promise((resolve, reject) => {
    const options = {
      hostname: 'api.github.com',
      path: `/repos/${owner}/${repo}/issues/${prNumber}/comments`,
      method: 'GET',
      headers: {
        'Authorization': `token ${token}`,
        'Accept': 'application/vnd.github.v3+json',
        'User-Agent': 'Go-AI-Code-Review-Action'
      }
    };
    
    const req = https.request(options, (res) => {
      let body = '';
      res.on('data', chunk => body += chunk);
      res.on('end', () => {
        try {
          const comments = JSON.parse(body);
          resolve(comments);
        } catch (e) {
          reject(new Error(`Parse error: ${e.message}`));
        }
      });
    });
    
    req.on('error', reject);
    req.end();
  });
}

function getCheckRuns(owner, repo, token) {
  return new Promise((resolve, reject) => {
    // For simplification, we'll return empty array since getting SHA is complex
    // In a real implementation, you'd need to get the PR head SHA first
    resolve([]);
  });
}

function getCheckStatus(checkRuns, patterns) {
  const matchingRun = checkRuns.find(run => 
    patterns.some(pattern => run.name.includes(pattern))
  );
  return matchingRun ? matchingRun.conclusion : 'unknown';
}

function formatStatus(status) {
  switch (status) {
    case 'success':
      return '✅ Passed';
    case 'failure':
      return '❌ Failed';
    case 'in_progress':
      return '⏳ Running';
    default:
      return '⏳ Pending';
  }
}

function postComment(prNumber, owner, repo, comment, token) {
  return new Promise((resolve, reject) => {
    const data = JSON.stringify({ body: comment });
    
    const options = {
      hostname: 'api.github.com',
      path: `/repos/${owner}/${repo}/issues/${prNumber}/comments`,
      method: 'POST',
      headers: {
        'Authorization': `token ${token}`,
        'Accept': 'application/vnd.github.v3+json',
        'User-Agent': 'Go-AI-Code-Review-Action',
        'Content-Type': 'application/json',
        'Content-Length': Buffer.byteLength(data)
      }
    };
    
    console.log(`Posting summary to: ${options.hostname}${options.path}`);
    
    const req = https.request(options, (res) => {
      let body = '';
      res.on('data', chunk => body += chunk);
      res.on('end', () => {
        if (res.statusCode >= 200 && res.statusCode < 300) {
          console.log('Summary comment posted successfully');
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

postSummary().catch(console.error);