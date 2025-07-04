// Usage:
//   bun run cicd.js --mode <dev|build>
//   bun run cicd.js --build                           # Cross-platform builds
//   bun run cicd.js [--commit] [--tag] [--push]       # Git operations
//   bun run cicd.js --release                         # Create GitHub release
//   bun run cicd.js --upload-r2                       # Upload to Cloudflare R2
//   bun run cicd.js --full-release                    # Complete GitHub release workflow
//   bun run cicd.js --r2-release                      # Complete R2 release workflow
//   bun run cicd.js --build --commit --tag --release  # Same as --full-release
// Usage: 2 modes (dev and build server) and deployment with git
import { $ } from "bun";
import path from "path";
import fs from "fs/promises";
import { parseArgs } from "util";

// --- Color Helpers ---
const reset = "\x1b[0m";
const colorFormat = "ansi";
const green = (text) => `${Bun.color("#2ecc71", colorFormat)}${text}${reset}`;
const red = (text) => `${Bun.color("#e74c3c", colorFormat)}${text}${reset}`;
const cyan = (text) => `${Bun.color("#3498db", colorFormat)}${text}${reset}`;
const yellow = (text) => `${Bun.color("#f1c40f", colorFormat)}${text}${reset}`;
const blue = cyan;
const bold = (text) => `\x1b[1m${text}${reset}`;

// --- Integrated Bun Spinner ---
// (Keep the createBunSpinner function exactly as it was)
function createBunSpinner(initialText = "", opts = {}) {
  const stream = opts.stream || process.stderr;
  const isTTY = !!stream.isTTY && process.env.TERM !== "dumb" && !process.env.CI;
  const frames = Array.isArray(opts.frames) && opts.frames.length
    ? opts.frames
    : isTTY
      ? ["⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"]
      : ["-"];
  const interval = opts.interval || 80;

  let text = initialText;
  let colorName = opts.color || "yellow";
  let colorFn = mapColor(colorName);
  let idx = 0;
  let timer = null;

  function mapColor(name) {
    switch ((name || "").toLowerCase()) {
      case "red": return red;
      case "green": return green;
      case "yellow": return yellow;
      case "blue": return cyan;
      case "cyan": return cyan;
      default: return (s) => s;
    }
  }

  function render() {
    const frame = frames[idx];
    if (isTTY) {
      // move to start of line, hide cursor, write frame + text
      stream.write("\r\x1b[?25l" + colorFn(frame) + " " + text);
    } else {
      // non-TTY: just print a line once
      stream.write(frame + " " + text + "\n");
    }
  }

  const spinner = {
    start(opts2 = {}) {
      const o = typeof opts2 === "string" ? { text: opts2 } : opts2;
      if (o.text != null) text = o.text;
      if (o.color) { colorName = o.color; colorFn = mapColor(colorName); }
      if (timer) clearInterval(timer);
      render();
      timer = setInterval(() => {
        idx = (idx + 1) % frames.length;
        render();
      }, interval);
      return spinner;
    },

    update(opts2 = {}) {
      const o = typeof opts2 === "string" ? { text: opts2 } : opts2;
      if (o.text != null) text = o.text;
      if (o.color) { colorName = o.color; colorFn = mapColor(colorName); }
      render();
      return spinner;
    },

    stop(opts2 = {}) {
      if (timer) { clearInterval(timer); timer = null; }
      const o = typeof opts2 === "string" ? { text: opts2 } : opts2;
      if (o.text != null) text = o.text;
      if (o.color) colorFn = mapColor(o.color);

      // pick a final mark
      const rawMark = o.mark != null ? o.mark : frames[idx];
      // assume o.mark is already colored if you passed green("✔"), etc.
      const mark = rawMark + " ";
      const out = mark + text + (isTTY ? "\n\x1b[?25h" : "\n");
      // overwrite the spinner line
      stream.write("\r" + out);
      return spinner;
    },

    success(opts2 = {}) {
      const o = typeof opts2 === "string" ? { text: opts2 } : opts2;
      return spinner.stop({
        mark: green("✔"),
        text: o.text != null ? o.text : text,
        color: "green"
      });
    },

    error(opts2 = {}) {
      const o = typeof opts2 === "string" ? { text: opts2 } : opts2;
      return spinner.stop({
        mark: red("✖"),
        text: o.text != null ? o.text : text,
        color: "red"
      });
    },

    warn(opts2 = {}) {
      const o = typeof opts2 === "string" ? { text: opts2 } : opts2;
      return spinner.stop({
        mark: yellow("⚠"),
        text: o.text != null ? o.text : text,
        color: "yellow"
      });
    },

    info(opts2 = {}) {
      const o = typeof opts2 === "string" ? { text: opts2 } : opts2;
      return spinner.stop({
        mark: cyan("ℹ"),
        text: o.text != null ? o.text : text,
        color: "cyan"
      });
    },

    isSpinning() {
      return timer != null;
    }
  };

  return spinner;
}
// --- End Integrated Bun Spinner ---


// --- Argument Parsing ---
const { values } = parseArgs({
  args: Bun.argv.slice(2), // Exclude 'bun' and script path
  options: {
    mode: { type: 'string' },
    build: { type: 'boolean', default: false },
    commit: { type: 'boolean', default: false },
    tag: { type: 'boolean', default: false },
    push: { type: 'boolean', default: false },
    release: { type: 'boolean', default: false },
    'upload-r2': { type: 'boolean', default: false },
    'full-release': { type: 'boolean', default: false },
    'r2-release': { type: 'boolean', default: false },
    help: { type: 'boolean', default: false },
  },
  strict: false,
  allowPositionals: false,
});

const mode = values.mode;
const isFullRelease = values['full-release'];
const isR2Release = values['r2-release'];
const showHelp = values.help;

// Check for help flag first
if (showHelp) {
  console.log(bold("Usage:"));
  console.log(cyan("  bun run cicd.js --mode <dev|build>"));
  console.log(cyan("  bun run cicd.js --build                           # Cross-platform builds"));
  console.log(cyan("  bun run cicd.js [--commit] [--tag] [--push]       # Git operations"));
  console.log(cyan("  bun run cicd.js --release                         # Create GitHub release"));
  console.log(cyan("  bun run cicd.js --upload-r2                       # Upload to Cloudflare R2"));
  console.log(cyan("  bun run cicd.js --full-release                    # Complete GitHub release workflow"));
  console.log(cyan("  bun run cicd.js --r2-release                      # Complete R2 release workflow"));
  console.log(cyan("  bun run cicd.js --build --commit --tag --release  # Same as --full-release"));
  console.log(cyan("  bun run cicd.js --help                            # Show this help"));
  process.exit(0);
}

// Set individual flags - override with full-release or r2-release if specified
const shouldBuild = values.build || isFullRelease || isR2Release;
const shouldCommit = values.commit || isFullRelease || isR2Release;
const shouldTag = values.tag || isFullRelease || isR2Release;
const shouldPush = values.push || isFullRelease || isR2Release;
const shouldRelease = values.release || isFullRelease;
const shouldUploadR2 = values['upload-r2'] || isR2Release;

// --- Helper Functions ---

/**
 * Generates a properly quoted -ldflags string so that it is treated as a
 * single argument by the shell.  Each key is written in -X=key=value form
 * to avoid accidental token splitting.
 *
 * @param {string} version
 * @param {string} buildDate
 * @param {string} gitCommit
 * @returns {string} ldflags eg. "-X=main.Version=0.1.3 -X=main.BuildDate=…"
 */
function generateLdflags(version, buildDate, gitCommit) {
  return `-X=main.Version=${version} -X=main.BuildDate=${buildDate} -X=main.GitCommit=${gitCommit}`;
}

/** Builds Go binaries for all supported platforms with version info */
async function buildCrossPlatform(version = null) {
  const platforms = [
    { os: 'darwin', arch: 'amd64', name: 'code4context-darwin-amd64' },
    { os: 'darwin', arch: 'arm64', name: 'code4context-darwin-arm64' },
    { os: 'linux', arch: 'amd64', name: 'code4context-linux-amd64' },
    { os: 'linux', arch: 'arm64', name: 'code4context-linux-arm64' },
    { os: 'windows', arch: 'amd64', name: 'code4context-windows-amd64.exe' },
    { os: 'windows', arch: 'arm64', name: 'code4context-windows-arm64.exe' }
  ];

  const mainSpinner = createBunSpinner(`🚀 Starting cross-platform Go builds...`).start();

  try {
    // Clean bin directory
    mainSpinner.update({ text: '🧹 Cleaning bin directory...' });
    await $`rm -rf bin/*`.nothrow();
    await $`mkdir -p bin`.throws(true);

    // Get build info
    const buildDate = new Date().toISOString();
    const gitCommitResult = await $`git rev-parse --short HEAD`.nothrow();
    const gitCommit = gitCommitResult.exitCode === 0 ? gitCommitResult.stdout.toString().trim() : 'unknown';
    const buildVersion = version || 'dev';

    // Build for each platform
    for (const platform of platforms) {
      mainSpinner.update({ text: `🔨 Building ${platform.name}...` });

      // Create ldflags for build info
      const ldflags = generateLdflags(buildVersion, buildDate, gitCommit);

      // Quote the ldflags string so it is passed as a single argument
      const buildResult = await $`go build -ldflags "${ldflags}" -o bin/${platform.name} .`
        .env({
          ...process.env,
          GOOS: platform.os,
          GOARCH: platform.arch,
          CGO_ENABLED: '0'
        })
        .nothrow();

      if (buildResult.exitCode !== 0) {
        mainSpinner.error({ text: red(`❌ Failed to build ${platform.name}`) });
        console.error(red(buildResult.stderr.toString()));
        throw new Error(`Build failed for ${platform.name}`);
      }
    }

    mainSpinner.success({ text: green(`✅ Built ${platforms.length} binaries successfully`) });

    // List built files
    console.log(cyan('📦 Built binaries:'));
    const lsResult = await $`ls -la bin/`.nothrow();
    if (lsResult.exitCode === 0) {
      console.log(lsResult.stdout.toString());
    }

  } catch (error) {
    if (mainSpinner.isSpinning()) {
      mainSpinner.error({ text: red('❌ Cross-platform build failed') });
    }
    throw error;
  }
}

/** Parses the latest entry from changelog.md. */
async function parseLatestChangelogEntry() {
  const changelogPath = path.join(import.meta.dir, 'changelog.md');
  console.log(cyan(`ℹ️ Reading changelog: ${changelogPath}`));
  try {
    const content = await fs.readFile(changelogPath, 'utf-8');
    const lines = content.split('\n');
    const headerStartRegex = /^#\s*([0-9]+\.[0-9]+\.[0-9]+[a-z]?)\s*-/i;
    const firstEntryStartIndex = lines.findIndex(line => headerStartRegex.test(line));
    if (firstEntryStartIndex === -1) throw new Error("Could not find any entry starting with '# <version> -' in changelog.md.");
    let nextEntryStartIndex = lines.findIndex((line, index) => index > firstEntryStartIndex && headerStartRegex.test(line));
    if (nextEntryStartIndex === -1) nextEntryStartIndex = lines.length;
    const entryLines = lines.slice(firstEntryStartIndex, nextEntryStartIndex);
    if (entryLines.length === 0) throw new Error("Detected an empty entry block in changelog.md.");
    const headerLine = entryLines[0];
    const headerMatch = headerLine.match(/^#\s*([0-9]+\.[0-9]+\.[0-9]+[a-z]?)\s*-\s*(.*)/i);
    if (!headerMatch || headerMatch.length < 3) throw new Error(`Could not parse header (line ${firstEntryStartIndex + 1}): "${headerLine}". Expected: '# <version> - <summary>'`);
    const version = headerMatch[1].trim();
    const summary = headerMatch[2].trim();
    const descriptionPoints = entryLines.slice(1).map(line => line.trim()).filter(line => line.startsWith('-')).map(line => `* ${line.substring(1).trim()}`).filter(line => line.length > 2);
    const description = descriptionPoints.join('\n');
    console.log(green(`✅ Parsed changelog: v${version} - ${summary}`));
    return { version, summary, description };
  } catch (error) {
    console.error(red(`❌ Error reading/parsing changelog.md: ${error.message}`));
    throw error;
  }
}

/** Checks if there are uncommitted changes. Throws error if clean. */
async function checkGitStatus() {
  console.log(cyan("ℹ️ Checking Git status..."));
  const { stdout, exitCode } = await $`git status --porcelain`.nothrow();
  if (exitCode !== 0) throw new Error(red("Failed to check Git status."));
  if (stdout.toString().trim() === '') throw new Error(red("❌ No changes detected. Working directory clean. Nothing to commit."));
  console.log(green("✅ Git status check passed: Changes detected."));
}

/** Checks if a Git tag already exists. Throws error if it does. */
async function checkGitTagExists(version) {
  const tagToCheck = `v${version}`;
  console.log(cyan(`ℹ️ Checking if Git tag ${tagToCheck} exists...`));
  const { stdout, exitCode } = await $`git tag -l ${tagToCheck}`.nothrow();
  if (exitCode === 0 && stdout.toString().trim() === tagToCheck) {
    throw new Error(red(`❌ Git tag '${tagToCheck}' already exists. Update changelog.md.`));
  }
  console.log(green(`✅ Git tag check passed: Tag '${tagToCheck}' does not exist yet.`));
}

/** Stages all changes. */
async function gitAdd() {
  const spinner = createBunSpinner(`ℹ️ Staging changes (${bold('git add .')})...`).start();
  try {
    await $`git add .`.quiet().throws(true);
    spinner.success({ text: green('✅ Changes staged.') });
  } catch (error) {
    spinner.error({ text: red('❌ Failed to stage changes.') });
    console.error(red(error.stderr?.toString() || error.message));
    throw error;
  }
}

/** Commits staged changes with a formatted message. */
async function gitCommit(summary, description) {
  const spinner = createBunSpinner(`ℹ️ Committing changes...`).start();
  const commitMessage = `${summary}\n\n${description}`;
  try {
    const proc = Bun.spawnSync(['git', 'commit', '--file=-'], { stdin: Buffer.from(commitMessage) });
    if (proc.exitCode !== 0) throw new Error(proc.stderr.toString() || "Git commit failed");
    spinner.success({ text: green('✅ Changes committed.') });
  } catch (error) {
    spinner.error({ text: red('❌ Git commit failed.') });
    console.error(red(error.message));
    throw error;
  }
}

/** Creates an annotated Git tag. */
async function gitTag(version, summary) {
  const tag = `v${version}`;
  const spinner = createBunSpinner(`ℹ️ Creating annotated Git tag ${bold(`'${tag}'`)}...`).start();
  try {
    await $`git tag -a ${tag} -m ${summary}`.quiet().throws(true);
    spinner.success({ text: green(`✅ Tag '${tag}' created.`) });
  } catch (error) {
    spinner.error({ text: red(`❌ Failed to create tag '${tag}'.`) });
    console.error(red(error.stderr?.toString() || error.message));
    throw error;
  }
}

/** Pushes commits and tags to the remote repository. */
async function gitPush() {
  const spinner = createBunSpinner(`ℹ️ Preparing to push...`).start();
  try {
    // Push commits
    spinner.update({ text: `ℹ️ Pushing commits (logs below)...`, color: 'cyan' });
    const pushCommitsResult = await $`git push`.nothrow();
    if (pushCommitsResult.exitCode !== 0) {
      spinner.error({ text: red(`❌ Failed to push commits.`) });
      console.error(red(pushCommitsResult.stderr.toString() || "git push failed"));
      throw new Error("Failed to push commits.");
    }
    spinner.success({ text: green(`✅ Commits pushed.`) });

    // Push tags
    spinner.start({ text: `ℹ️ Pushing tags (logs below)...`, color: 'cyan' });
    const pushTagsResult = await $`git push --tags`.nothrow();
    if (pushTagsResult.exitCode !== 0) {
      spinner.error({ text: red(`❌ Failed to push tags.`) });
      console.error(red(pushTagsResult.stderr.toString() || "git push --tags failed"));
      throw new Error("Failed to push tags.");
    }
    spinner.success({ text: green(`✅ Tags pushed.`) });

  } catch (error) {
    if (spinner.isSpinning()) {
      spinner.error({ text: red(`❌ Push operation failed.`) });
    }
    throw error;
  }
}

/** Builds the main Go binary for current platform */
async function buildLocal() {
  const spinner = createBunSpinner(`🚀 Building code4context for current platform...`).start();

  try {
    // Get build info
    const buildDate = new Date().toISOString();
    const gitCommitResult = await $`git rev-parse --short HEAD`.nothrow();
    const gitCommit = gitCommitResult.exitCode === 0 ? gitCommitResult.stdout.toString().trim() : 'unknown';

    // Create ldflags for build info
    const ldflags = generateLdflags('dev', buildDate, gitCommit);

    // Quote the ldflags string so it is passed as a single argument
    const buildResult = await $`go build -ldflags "${ldflags}" -o code4context .`
      .env({ ...process.env, CGO_ENABLED: '0' })
      .nothrow();

    if (buildResult.exitCode !== 0) {
      spinner.error({ text: red('❌ Local build failed') });
      console.error(red(buildResult.stderr.toString()));
      throw new Error('Local build failed');
    }

    spinner.success({ text: green('✅ Local build completed') });

  } catch (error) {
    if (spinner.isSpinning()) {
      spinner.error({ text: red('❌ Local build process failed') });
    }
    throw error;
  }
}

/** Calculates SHA256 hash of a file */
async function calculateFileHash(filePath) {
  const hasher = new Bun.CryptoHasher("sha256");
  const file = Bun.file(filePath);
  const arrayBuffer = await file.arrayBuffer();
  hasher.update(new Uint8Array(arrayBuffer));
  return hasher.digest("hex");
}

/** Checks if a binary with the same hash already exists in R2 */
async function checkExistingBinary(hash, awsEnv, bucket, endpoint) {
  const binaryPath = `binaries/code4context-${hash}`;
  const checkResult = await $`aws s3 ls s3://${bucket}/${binaryPath} --endpoint-url ${endpoint}`
    .env(awsEnv)
    .nothrow();
  
  return checkResult.exitCode === 0;
}

/** Uploads binaries to Cloudflare R2 with hash-based deduplication */
async function uploadToR2(version) {
  const spinner = createBunSpinner(`☁️ Uploading binaries to Cloudflare R2...`).start();

  try {
    // Check required environment variables
    const requiredEnvVars = ['R2_ACCESS_KEY_ID', 'R2_SECRET_ACCESS_KEY', 'R2_BUCKET_NAME', 'R2_ENDPOINT'];
    const missingVars = requiredEnvVars.filter(varName => !process.env[varName]);

    if (missingVars.length > 0) {
      throw new Error(`Missing required environment variables: ${missingVars.join(', ')}`);
    }

    // Check if AWS CLI is available
    const awsCheckResult = await $`which aws`.nothrow();
    if (awsCheckResult.exitCode !== 0) {
      throw new Error('AWS CLI is not installed. Please install it first.');
    }

    const bucket = process.env.R2_BUCKET_NAME;
    const endpoint = process.env.R2_ENDPOINT;
    const versionPath = `releases/v${version}`;

    // Configure AWS CLI for R2
    spinner.update({ text: `🔧 Configuring AWS CLI for R2...` });

    const awsEnv = {
      ...process.env,
      AWS_ACCESS_KEY_ID: process.env.R2_ACCESS_KEY_ID,
      AWS_SECRET_ACCESS_KEY: process.env.R2_SECRET_ACCESS_KEY,
      AWS_DEFAULT_REGION: 'auto'
    };

    // Get binary files
    const binFiles = await $`ls bin/`.env(awsEnv).nothrow();
    if (binFiles.exitCode !== 0) {
      throw new Error('No binaries found in bin/ directory');
    }

    const files = binFiles.stdout.toString().trim().split('\n').filter(f => f.trim());
    const metadata = {};

    // Process each binary with hash-based deduplication
    for (const file of files) {
      spinner.update({ text: `🔍 Processing ${file}...` });

      const filePath = `bin/${file}`;
      const hash = await calculateFileHash(filePath);
      const platform = file.replace('code4context-', '').replace('.exe', '');
      
      // Check if binary with this hash already exists
      const binaryExists = await checkExistingBinary(hash, awsEnv, bucket, endpoint);
      
      if (binaryExists) {
        spinner.update({ text: `♻️  Reusing existing binary for ${file} (hash: ${hash.substring(0, 8)}...)` });
        
        // Create metadata entry pointing to existing binary
        const baseUrl = process.env.R2_PUBLIC_URL || `https://${bucket}.${endpoint.replace('https://', '')}`;
        metadata[platform] = {
          hash: hash,
          binary_url: `${baseUrl}/binaries/code4context-${hash}`,
          reused: true
        };
      } else {
        spinner.update({ text: `📤 Uploading new binary ${file}...` });
        
        // Upload to hash-based location
        const hashBinaryPath = `binaries/code4context-${hash}`;
        const uploadResult = await $`aws s3 cp ${filePath} s3://${bucket}/${hashBinaryPath} --endpoint-url ${endpoint}`
          .env(awsEnv)
          .nothrow();

        if (uploadResult.exitCode !== 0) {
          throw new Error(`Failed to upload ${file}: ${uploadResult.stderr.toString()}`);
        }

        // Create metadata entry
        const baseUrl = process.env.R2_PUBLIC_URL || `https://${bucket}.${endpoint.replace('https://', '')}`;
        metadata[platform] = {
          hash: hash,
          binary_url: `${baseUrl}/binaries/code4context-${hash}`,
          reused: false
        };
      }

      // Also upload to version-specific path for backward compatibility
      spinner.update({ text: `📋 Creating version-specific reference for ${file}...` });
      const versionUploadResult = await $`aws s3 cp ${filePath} s3://${bucket}/${versionPath}/${file} --endpoint-url ${endpoint}`
        .env(awsEnv)
        .nothrow();

      if (versionUploadResult.exitCode !== 0) {
        throw new Error(`Failed to upload version-specific ${file}: ${versionUploadResult.stderr.toString()}`);
      }
    }

    // Create and upload metadata.json
    spinner.update({ text: `📋 Creating metadata.json...` });

    const metadataContent = {
      version: version,
      created_at: new Date().toISOString(),
      binaries: metadata
    };

    const metadataFile = `metadata-${version}.json`;
    await fs.writeFile(metadataFile, JSON.stringify(metadataContent, null, 2));

    const metadataUploadResult = await $`aws s3 cp ${metadataFile} s3://${bucket}/${versionPath}/metadata.json --endpoint-url ${endpoint}`
      .env(awsEnv)
      .nothrow();

    if (metadataUploadResult.exitCode !== 0) {
      throw new Error(`Failed to upload metadata: ${metadataUploadResult.stderr.toString()}`);
    }

    // Upload latest version marker
    spinner.update({ text: `📝 Updating latest version marker...` });

    const versionFile = `latest-version.txt`;
    await fs.writeFile(versionFile, version);

    const versionUploadResult = await $`aws s3 cp ${versionFile} s3://${bucket}/${versionFile} --endpoint-url ${endpoint}`
      .env(awsEnv)
      .nothrow();

    if (versionUploadResult.exitCode !== 0) {
      throw new Error(`Failed to upload version marker: ${versionUploadResult.stderr.toString()}`);
    }

    // Upload install script
    spinner.update({ text: `📜 Uploading install script...` });

    const installScriptResult = await $`aws s3 cp install.sh s3://${bucket}/install.sh --endpoint-url ${endpoint}`
      .env(awsEnv)
      .nothrow();

    if (installScriptResult.exitCode !== 0) {
      throw new Error(`Failed to upload install script: ${installScriptResult.stderr.toString()}`);
    }

    // Clean up temp files
    await $`rm ${versionFile} ${metadataFile}`.nothrow();

    spinner.success({ text: green(`✅ Binaries uploaded to R2 successfully`) });

    // Show download URLs and reuse statistics
    const baseUrl = process.env.R2_PUBLIC_URL || `https://${bucket}.${endpoint.replace('https://', '')}`;
    console.log(cyan(`🔗 Binaries available at: ${baseUrl}/${versionPath}/`));
    console.log(cyan(`🔗 Metadata available at: ${baseUrl}/${versionPath}/metadata.json`));
    console.log(cyan(`🔗 Install script available at: ${baseUrl}/install.sh`));

    // Show reuse statistics
    const reusedCount = Object.values(metadata).filter(m => m.reused).length;
    const newCount = Object.values(metadata).filter(m => !m.reused).length;
    console.log(cyan(`♻️  Binary reuse: ${reusedCount} reused, ${newCount} new`));

    // List uploaded files
    console.log(cyan('📦 Version-specific files:'));
    files.forEach(file => {
      console.log(`  ${baseUrl}/${versionPath}/${file}`);
    });
    console.log(`  ${baseUrl}/${versionPath}/metadata.json`);
    console.log(`  ${baseUrl}/install.sh`);

    // List optimized binary URLs
    console.log(cyan('🚀 Optimized binary URLs:'));
    Object.entries(metadata).forEach(([platform, meta]) => {
      const status = meta.reused ? '♻️  (reused)' : '🆕 (new)';
      console.log(`  ${platform}: ${meta.binary_url} ${status}`);
    });

  } catch (error) {
    if (spinner.isSpinning()) {
      spinner.error({ text: red('❌ R2 upload failed') });
    }
    throw error;
  }
}

/** Creates a GitHub release and uploads binaries */
async function createGitHubRelease(version, summary, description) {
  const spinner = createBunSpinner(`🚀 Creating GitHub release v${version}...`).start();

  try {
    // Check if gh CLI is available
    const ghCheckResult = await $`which gh`.nothrow();
    if (ghCheckResult.exitCode !== 0) {
      throw new Error('GitHub CLI (gh) is not installed. Please install it first.');
    }

    // Create the release
    spinner.update({ text: `📦 Creating release v${version}...` });

    const releaseBody = `${summary}\n\n## Changes\n${description}`;
    const createResult = await $`gh release create v${version} --title "v${version}" --notes ${releaseBody}`.nothrow();

    if (createResult.exitCode !== 0) {
      throw new Error(`Failed to create GitHub release: ${createResult.stderr.toString()}`);
    }

    // Upload binaries
    spinner.update({ text: `📤 Uploading binaries...` });

    const uploadResult = await $`gh release upload v${version} bin/*`.nothrow();

    if (uploadResult.exitCode !== 0) {
      throw new Error(`Failed to upload binaries: ${uploadResult.stderr.toString()}`);
    }

    spinner.success({ text: green(`✅ GitHub release v${version} created successfully`) });

    // Show release URL
    const releaseUrl = `https://github.com/jasonwillschiu/code4context-com/releases/tag/v${version}`;
    console.log(cyan(`🔗 Release URL: ${releaseUrl}`));

  } catch (error) {
    if (spinner.isSpinning()) {
      spinner.error({ text: red('❌ GitHub release creation failed') });
    }
    throw error;
  }
}




// --- Main Logic ---

// Dev Mode
if (mode === 'dev') {
  console.log(cyan("🚀 Building frontend and starting Go backend..."));

  const buildSpinner = createBunSpinner(`🚀 Building Astro frontend...`).start();
  try {
    // Build frontend
    buildSpinner.update({ text: `🚀 Building frontend in ${bold('frontend')}...` });
    const frontendBuildResult = await $`cd frontend && bunx --bun astro build`
      .env({ ...process.env, FORCE_COLOR: '1' })
      .nothrow();

    if (frontendBuildResult.exitCode !== 0) {
      buildSpinner.error({ text: red('❌ Frontend build failed.') });
      console.error(red(frontendBuildResult.stderr.toString()));
      process.exit(1);
    }

    // Copy frontend build to backend
    buildSpinner.update({ text: `📁 Copying frontend build to backend/mpa/...` });
    await $`rm -rf backend/mpa && mkdir -p backend/mpa`.quiet().throws(true);
    await $`cp -r frontend/dist/* backend/mpa/`.quiet().throws(true);

    buildSpinner.success({ text: green('✅ Frontend build and copy completed.') });

    // Start Go backend
    console.log(green("🟢 Starting Go backend with embedded frontend (localhost:8099)..."));
    console.log(yellow("⏳ Press Ctrl+C to stop the server."));

    const backendResult = await $`cd backend && go run main.go`
      .env({ ...process.env, FORCE_COLOR: '1' })
      .nothrow();

    if (backendResult.exitCode !== 0 && backendResult.signal !== 'SIGINT' && backendResult.signal !== 'SIGTERM') {
      console.error(red(`Go server exited unexpectedly: Code ${backendResult.exitCode} Signal ${backendResult.signal}`));
    }

  } catch (error) {
    if (buildSpinner.isSpinning()) {
      buildSpinner.error({ text: red('❌ Build process failed.') });
    }
    console.error(red("🚨 Error during dev build:"), error);
    process.exit(1);
  } finally {
    console.log(yellow("\n🛑 Shutting down development environment..."));
  }
  console.log(green("✅ Development environment stopped."));

  // Build Mode - Local binary only
} else if (mode === 'build') {
  console.log(cyan("🚀 Building local MCP server binary..."));

  try {
    await buildLocal();
    console.log(green("✅ Local build completed successfully."));
    console.log(cyan("📁 Binary available at: ./code4context"));
    console.log(yellow("ℹ️ Run with: ./code4context"));

  } catch (error) {
    console.error(red("🚨 Error during build:"), error);
    process.exit(1);
  }

  // CICD / Build Mode
} else if (shouldBuild || shouldCommit || shouldTag || shouldPush || shouldRelease || shouldUploadR2) {

  console.log(cyan("🚀 Starting CICD actions..."));
  let changelogData;

  try {
    if (shouldCommit || shouldTag || shouldPush || shouldRelease || shouldBuild || shouldUploadR2) {
      changelogData = await parseLatestChangelogEntry();
      const { version, summary, description } = changelogData;

      if (shouldBuild) {
        await buildCrossPlatform(version);
      }

      if (shouldCommit) {
        await checkGitStatus();
        await gitAdd();
        await gitCommit(summary, description);
      }

      if (shouldTag) {
        await checkGitTagExists(version);
        await gitTag(version, summary);
      }

      if (shouldPush) {
        await gitPush();
      }

      if (shouldRelease) {
        await createGitHubRelease(version, summary, description);
      }

      if (shouldUploadR2) {
        await uploadToR2(version);
      }
    }

    console.log(green("\n✅ CICD actions completed successfully."));

  } catch (error) {
    console.error(red(`\n🚨 CICD process failed: ${error.message}`));
    if (error.stack && !error.message.startsWith('❌') && !error.message.includes('failed.')) {
      console.error(error.stack);
    }
    process.exit(1);
  }

} else {
  console.error(red("❌ Invalid or missing mode/action specified."));
  console.log(bold("Usage:"));
  console.log(cyan("  bun run cicd.js --mode <dev|build>"));
  console.log(cyan("  bun run cicd.js --build                           # Cross-platform builds"));
  console.log(cyan("  bun run cicd.js [--commit] [--tag] [--push]       # Git operations"));
  console.log(cyan("  bun run cicd.js --release                         # Create GitHub release"));
  console.log(cyan("  bun run cicd.js --upload-r2                       # Upload to Cloudflare R2"));
  console.log(cyan("  bun run cicd.js --full-release                    # Complete GitHub release workflow"));
  console.log(cyan("  bun run cicd.js --r2-release                      # Complete R2 release workflow"));
  console.log(cyan("  bun run cicd.js --build --commit --tag --release  # Same as --full-release")); process.exit(1);
}
