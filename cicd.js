// Usage:
//   bun run cicd.js --mode <dev|build>
//   bun run cicd.js --build                           # Cross-platform builds
//   bun run cicd.js [--commit] [--tag] [--push]       # Git operations
//   bun run cicd.js --release                         # Create GitHub release
//   bun run cicd.js --upload-r2                       # Upload to Cloudflare R2
//   bun run cicd.js --full-release                    # Complete GitHub release workflow
//   bun run cicd.js --r2-release                      # Complete R2 release workflow
//   bun run cicd.js --build --commit --tag --release  # Same as --full-release
//   bun run cicd.js --install-guide                   # Show installation guide for AI tools
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
    'install-guide': { type: 'boolean', default: false },
  },
  strict: false,
  allowPositionals: false,
});

const mode = values.mode;
const isFullRelease = values['full-release'];
const isR2Release = values['r2-release'];
const showHelp = values.help;
const showInstallGuide = values['install-guide'];

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
  console.log(cyan("  bun run cicd.js --install-guide                   # Show installation guide for AI tools"));
  console.log(cyan("  bun run cicd.js --help                            # Show this help"));
  process.exit(0);
}

// Check for install guide flag
if (showInstallGuide) {
  showInstallationGuide();
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
    { os: 'windows', arch: 'amd64', name: 'code4context-windows-amd64.exe' }
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

/** Checks if any .go files are staged or modified */
async function hasGoFileChanges() {
  const { stdout, exitCode } = await $`git status --porcelain`.nothrow();
  if (exitCode !== 0) return true; // If we can't check, assume changes

  const changes = stdout.toString().trim().split('\n').filter(line => line.trim());
  const goFileChanges = changes.filter(line => line.includes('.go'));

  return goFileChanges.length > 0;
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

/** Calculates content hash based on Go source files that affect the binary */
async function calculateContentHash() {
  // Hash only the files that actually affect the Go binary compilation
  // This excludes README, docs, install scripts, etc.

  const hasher = new Bun.CryptoHasher("sha256");

  // Get list of Go source files
  const goFilesResult = await $`find . -name "*.go" -not -path "./test-files/*" | sort`.nothrow();
  if (goFilesResult.exitCode === 0) {
    const goFiles = goFilesResult.stdout.toString().trim().split('\n').filter(f => f.trim());

    for (const file of goFiles) {
      try {
        const content = await fs.readFile(file.trim(), 'utf-8');
        hasher.update(new TextEncoder().encode(file));
        hasher.update(new TextEncoder().encode(content));
      } catch (e) {
        // Skip files that can't be read
      }
    }
  }

  // Hash go.mod for dependency changes
  try {
    const goMod = await fs.readFile('go.mod', 'utf-8');
    hasher.update(new TextEncoder().encode('go.mod'));
    hasher.update(new TextEncoder().encode(goMod));
  } catch (e) {
    // Ignore if file doesn't exist
  }

  // Hash go.sum for dependency lock changes
  try {
    const goSum = await fs.readFile('go.sum', 'utf-8');
    hasher.update(new TextEncoder().encode('go.sum'));
    hasher.update(new TextEncoder().encode(goSum));
  } catch (e) {
    // Ignore if file doesn't exist
  }

  return hasher.digest("hex");
}

/** Gets the latest version's metadata to check for hash matches */
async function getLatestVersionMetadata(awsEnv, bucket, endpoint) {
  try {
    // Get latest version
    const latestResult = await $`aws s3 cp s3://${bucket}/latest-version.txt - --endpoint-url ${endpoint}`
      .env(awsEnv)
      .nothrow();

    if (latestResult.exitCode !== 0) {
      return null; // No previous version
    }

    const latestVersion = latestResult.stdout.toString().trim();

    // Get metadata for latest version
    const metadataResult = await $`aws s3 cp s3://${bucket}/releases/v${latestVersion}/metadata.json - --endpoint-url ${endpoint}`
      .env(awsEnv)
      .nothrow();

    if (metadataResult.exitCode !== 0) {
      return null; // No metadata found
    }

    return {
      version: latestVersion,
      metadata: JSON.parse(metadataResult.stdout.toString())
    };
  } catch (e) {
    return null;
  }
}

/** Checks if a binary actually exists in R2 for a given version and platform */
async function checkBinaryExists(awsEnv, bucket, endpoint, version, platform) {
  try {
    const fileName = `code4context-${platform}`;
    const checkResult = await $`aws s3api head-object --bucket ${bucket} --key releases/v${version}/${fileName} --endpoint-url ${endpoint}`
      .quiet()
      .env(awsEnv)
      .nothrow();

    return checkResult.exitCode === 0;
  } catch (e) {
    return false;
  }
}

/** Finds the latest version that actually has binaries for a given platform */
async function findLatestVersionWithBinary(awsEnv, bucket, endpoint, platform, _ignored = 0) {
  try {
    const objectPattern = `releases/`;
    const listCmd = $`aws s3 ls s3://${bucket}/${objectPattern} --recursive --endpoint-url ${endpoint}`
      .env(awsEnv)
      .nothrow();

    const listResult = await listCmd;
    if (listResult.exitCode !== 0) {
      return null;
    }

    const keyRegex = new RegExp(`releases/v([0-9]+\\.[0-9]+\\.[0-9]+[a-z]?)\\/code4context-${platform.replace('.', '\\.')}$`);
    /** @type {Set<string>} */
    const versions = new Set();

    for (const line of listResult.stdout.toString().split('\n')) {
      const trimmed = line.trim();
      if (!trimmed) continue;

      // Example line format (aws s3 ls):
      // 2025-07-04 14:33:07    5957442 releases/v0.4.8/code4context-darwin-arm64
      const match = trimmed.match(keyRegex);
      if (match && match[1]) {
        versions.add(match[1]);
      }
    }

    if (versions.size === 0) {
      return null;
    }

    // Convert set to array and sort descending semver
    const sorted = Array.from(versions).sort((a, b) => {
      const semverToInts = (v) =>
        v.split('.').map((part) => {
          // Strip any trailing letters (e.g., 1.2.3a)
          const num = part.replace(/[^\d]/g, '');
          return parseInt(num || '0', 10);
        });

      const aParts = semverToInts(a);
      const bParts = semverToInts(b);

      for (let i = 0; i < Math.max(aParts.length, bParts.length); i++) {
        const diff = (bParts[i] || 0) - (aParts[i] || 0);
        if (diff !== 0) return diff;
      }
      return 0;
    });

    return sorted[0] || null;
  } catch {
    return null;
  }
}

/**
 * Uploads binaries to Cloudflare R2 with smart reuse from previous versions.
 *
 * @param {string}  version
 * @param {boolean} [skipBuild=false]           If true, skip building; always reuse binaries.
 * @param {string}  [releaseSummary=null]       Short summary from changelog.
 * @param {string}  [releaseDescription=null]   Long description / bullet list from changelog.
 *
 * OPTIMIZATION: Instead of copying unchanged binaries to each new version folder,
 * this function maintains a global binary-mapping.json that tracks where each
 * platform's binary actually exists. This saves significant R2 storage space
 * and speeds up deployment by avoiding redundant uploads.
 *
 * SIMPLE APPROACH: binary-mapping.json is the source of truth showing where
 * binaries actually exist. metadata.json now also carries release notes to make
 * it a more useful artifact for consumers / dashboards.
 */
async function uploadToR2(version, skipBuild = false, releaseSummary = null, releaseDescription = null) {
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

    // Calculate current content hash
    spinner.update({ text: `🔍 Calculating content hash...` });
    const currentHash = await calculateContentHash();

    // Check if we can reuse binaries from the latest version
    spinner.update({ text: `🔍 Checking for reusable binaries...` });
    const latestVersionData = await getLatestVersionMetadata(awsEnv, bucket, endpoint);

    const platforms = [
      'darwin-amd64', 'darwin-arm64', 'linux-amd64',
      'linux-arm64', 'windows-amd64.exe'
    ];

    let canReuseAll = false;
    let sourceVersion = null;

    // If we skipped building, we definitely want to reuse
    if (skipBuild) {
      canReuseAll = true;
      if (latestVersionData) {
        sourceVersion = latestVersionData.version;
        spinner.update({ text: `♻️  No .go files changed, reusing all binaries from v${sourceVersion}` });
      } else {
        throw new Error('No previous version found to reuse binaries from');
      }
    } else if (latestVersionData && latestVersionData.metadata.content_hash === currentHash) {
      canReuseAll = true;
      sourceVersion = latestVersionData.version;
      spinner.update({ text: `♻️  Content unchanged, reusing all binaries from v${sourceVersion}` });
    }

    const metadata = {
      version: version,
      created_at: new Date().toISOString(),
      content_hash: currentHash,
      release_summary: releaseSummary,
      release_description: releaseDescription,
      binaries: {},
      binary_source_versions: {}
    };

    if (canReuseAll) {
      // Fast path: Reference existing binaries instead of copying
      for (const platform of platforms) {
        const fileName = `code4context-${platform}`;
        spinner.update({ text: `🔗 Referencing ${fileName} from v${sourceVersion}...` });

        // For metadata.json: store reference info (can show immediate source)
        const baseUrl = process.env.R2_PUBLIC_URL || `https://${bucket}.${endpoint.replace('https://', '')}`;
        metadata.binaries[platform] = {
          url: `${baseUrl}/releases/v${sourceVersion}/${fileName}`,
          reused_from: sourceVersion,
          last_updated_version: sourceVersion
        };
        metadata.binary_source_versions[platform] = sourceVersion;
      }
    } else {
      // Need to build and upload new binaries
      const binFiles = await $`ls bin/`.env(awsEnv).nothrow();
      if (binFiles.exitCode !== 0) {
        throw new Error('No binaries found in bin/ directory');
      }

      const files = binFiles.stdout.toString().trim().split('\n').filter(f => f.trim());

      for (const file of files) {
        spinner.update({ text: `📤 Uploading ${file}...` });

        const filePath = `bin/${file}`;
        const uploadResult = await $`aws s3 cp ${filePath} s3://${bucket}/${versionPath}/${file} --endpoint-url ${endpoint}`
          .env(awsEnv)
          .nothrow();

        if (uploadResult.exitCode !== 0) {
          throw new Error(`Failed to upload ${file}: ${uploadResult.stderr.toString()}`);
        }

        const platform = file.replace('code4context-', '');
        const baseUrl = process.env.R2_PUBLIC_URL || `https://${bucket}.${endpoint.replace('https://', '')}`;
        metadata.binaries[platform] = {
          url: `${baseUrl}/${versionPath}/${file}`,
          newly_built: true,
          last_updated_version: version
        };
        metadata.binary_source_versions[platform] = version;
      }
    }

    // Upload version-specific metadata.json
    // This contains detailed info about this specific version including where each binary comes from
    spinner.update({ text: `📋 Creating metadata.json...` });
    const metadataFile = `metadata-${version}.json`;
    await fs.writeFile(metadataFile, JSON.stringify(metadata, null, 2));

    const metadataUploadResult = await $`aws s3 cp ${metadataFile} s3://${bucket}/${versionPath}/metadata.json --endpoint-url ${endpoint}`
      .env(awsEnv)
      .nothrow();

    if (metadataUploadResult.exitCode !== 0) {
      throw new Error(`Failed to upload metadata: ${metadataUploadResult.stderr.toString()}`);
    }

    // Create/update global binary-mapping.json
    // This is the source of truth that tells install.sh where to find each platform's binary
    spinner.update({ text: `📋 Updating binary mapping with verification...` });
    const binaryMappingFile = `binary-mapping.json`;

    // Get existing mapping or create new one
    let globalMapping = {};
    const existingMappingResult = await $`aws s3 cp s3://${bucket}/binary-mapping.json - --endpoint-url ${endpoint}`
      .env(awsEnv)
      .nothrow();

    if (existingMappingResult.exitCode === 0) {
      try {
        globalMapping = JSON.parse(existingMappingResult.stdout.toString());
      } catch (e) {
        // If parsing fails, start with empty mapping
        globalMapping = {};
      }
    }

    // Update mapping with current version info
    globalMapping.last_updated = new Date().toISOString();
    globalMapping.latest_version = version;
    if (!globalMapping.binary_sources) {
      globalMapping.binary_sources = {};
    }

    // For each platform, determine where the binary actually exists
    for (const platform of platforms) {
      if (canReuseAll) {
        // If reusing, verify the existing mapping points to a real binary
        const existingSource = globalMapping.binary_sources[platform];
        if (existingSource) {
          // Verify the existing source actually has the binary
          const binaryExists = await checkBinaryExists(awsEnv, bucket, endpoint, existingSource, platform);
          if (binaryExists) {
            // Keep pointing to the verified location
            globalMapping.binary_sources[platform] = existingSource;
          } else {
            // Find the latest version that actually has this binary
            spinner.update({ text: `🔍 Finding latest version with ${platform} binary...` });
            const actualSource = await findLatestVersionWithBinary(awsEnv, bucket, endpoint, platform);
            if (actualSource) {
              globalMapping.binary_sources[platform] = actualSource;
              console.log(green(`✅ Found ${platform} binary in v${actualSource}`));
            } else {
              // Fallback to source version if no existing binary found
              globalMapping.binary_sources[platform] = sourceVersion;
              console.log(yellow(`⚠️  No existing ${platform} binary found, using v${sourceVersion}`));
            }
          }
        } else {
          // No existing mapping, find the latest version with this binary
          spinner.update({ text: `🔍 Finding latest version with ${platform} binary...` });
          const actualSource = await findLatestVersionWithBinary(awsEnv, bucket, endpoint, platform);
          if (actualSource) {
            globalMapping.binary_sources[platform] = actualSource;
            console.log(green(`✅ Found ${platform} binary in v${actualSource}`));
          } else {
            // Fallback to source version if no existing binary found
            globalMapping.binary_sources[platform] = sourceVersion;
            console.log(yellow(`⚠️  No existing ${platform} binary found, using v${sourceVersion}`));
          }
        }
      } else {
        // New binaries are uploaded to current version
        globalMapping.binary_sources[platform] = version;
      }
    }

    // Save to local repo for git commit
    await fs.writeFile(binaryMappingFile, JSON.stringify(globalMapping, null, 2));

    // Upload to R2
    const mappingUploadResult = await $`aws s3 cp ${binaryMappingFile} s3://${bucket}/binary-mapping.json --endpoint-url ${endpoint}`
      .env(awsEnv)
      .nothrow();

    if (mappingUploadResult.exitCode !== 0) {
      throw new Error(`Failed to upload binary mapping: ${mappingUploadResult.stderr.toString()}`);
    }

    // Update latest version marker
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

    // Clean up temp files (keep binary-mapping.json for git commit)
    await $`rm ${versionFile} ${metadataFile}`.nothrow();

    spinner.success({ text: green(`✅ Binaries uploaded to R2 successfully`) });

    // Show results
    const baseUrl = process.env.R2_PUBLIC_URL || `https://${bucket}.${endpoint.replace('https://', '')}`;
    console.log(cyan(`🔗 Release available at: ${baseUrl}/${versionPath}/`));
    console.log(cyan(`🔗 Install script: ${baseUrl}/install.sh`));
    console.log(cyan(`📋 Binary mapping saved to: ${binaryMappingFile} (ready for git commit)`));

    if (canReuseAll) {
      console.log(cyan(`♻️  All binaries reused from v${sourceVersion} (content hash: ${currentHash.substring(0, 8)}...)`));
    } else {
      console.log(cyan(`🆕 All binaries newly built (content hash: ${currentHash.substring(0, 8)}...)`));
    }

  } catch (error) {
    if (spinner.isSpinning()) {
      spinner.error({ text: red('❌ R2 upload failed') });
    }
    throw error;
  }
}

/** Shows installation guide for AI coding tools */
function showInstallationGuide() {
  console.log(bold(green("🚀 Code4Context Installation Guide for AI Coding Tools")));
  console.log("");

  // Installation first
  console.log(bold(cyan("📦 Step 1: Install Code4Context Binary")));
  console.log("");
  console.log(yellow("Quick install (recommended):"));
  console.log(`  ${green("curl -fsSL https://raw.githubusercontent.com/jasonwillschiu/code4context-com/main/install.sh | bash")}`);
  console.log("");
  console.log(yellow("Or download manually from:"));
  console.log(`  ${green("https://github.com/jasonwillschiu/code4context-com/releases")}`);
  console.log("");

  // Configuration for each tool
  console.log(bold(cyan("📱 Step 2: Configure Your AI Tool")));
  console.log("");

  // OpenCode
  console.log(bold(yellow("OpenCode:")));
  console.log(yellow("Create or edit your opencode.json file:"));
  console.log(green(`{
  "$schema": "https://opencode.ai/config.json",
  "mcp": {
    "code4context": {
      "type": "local",
      "command": [
        "/path/to/code4context"
      ],
      "environment": {}
    }
  }
}`));
  console.log("");

  // Claude Code
  console.log(bold(yellow("Claude Code:")));
  console.log(yellow("Add the MCP server using the binary path:"));
  console.log(`  ${green("claude mcp add code4context -- /path/to/code4context")}`);
  console.log("");
  console.log(yellow("For HTTP transport:"));
  console.log(`  ${green("claude mcp add --transport http code4context https://your-server.com/mcp")}`);
  console.log("");
  console.log(yellow("For SSE transport:"));
  console.log(`  ${green("claude mcp add --transport sse code4context https://your-server.com/sse")}`);
  console.log("");

  // Cursor
  console.log(bold(yellow("Cursor:")));
  console.log(yellow("Create or edit your mcp.json file:"));
  console.log(green(`{
  "mcpServers": {
    "code4context": {
      "command": "/path/to/code4context",
      "args": []
    }
  }
}`));
  console.log("");
  console.log(yellow("For remote server:"));
  console.log(green(`{
  "mcpServers": {
    "code4context": {
      "url": "https://your-server.com/mcp"
    }
  }
}`));
  console.log("");

  // Usage tips
  console.log(bold(cyan("💡 Usage Tips")));
  console.log("");
  console.log(yellow("• The binary generates structured code summaries (codebrev.md)"));
  console.log(yellow("• Works with any codebase - Go, JavaScript, TypeScript, Python, etc."));
  console.log(yellow("• Provides function signatures, types, and file organization"));
  console.log(yellow("• Use in prompts: 'analyze this codebase using code4context'"));
  console.log("");

  // Binary usage
  console.log(bold(cyan("🔧 Binary Usage")));
  console.log("");
  console.log(yellow("Generate code summary for current directory:"));
  console.log(`  ${green("./code4context")}`);
  console.log("");
  console.log(yellow("Generate summary for specific directory:"));
  console.log(`  ${green("./code4context /path/to/project")}`);
  console.log("");
  console.log(yellow("Check version:"));
  console.log(`  ${green("./code4context --version")}`);
  console.log("");

  // Path examples
  console.log(bold(cyan("📍 Common Installation Paths")));
  console.log("");
  console.log(yellow("If installed to current directory:"));
  console.log(`  ${green("./code4context")}`);
  console.log("");
  console.log(yellow("If installed to ~/.local/bin:"));
  console.log(`  ${green("~/.local/bin/code4context")}`);
  console.log("");
  console.log(yellow("If installed to /usr/local/bin:"));
  console.log(`  ${green("/usr/local/bin/code4context")}`);
  console.log("");
  console.log(yellow("Or if in PATH, just:"));
  console.log(`  ${green("code4context")}`);
  console.log("");

  console.log(bold(green("✅ Ready to enhance your AI coding experience!")));
  console.log(cyan("📚 For more details, visit: https://github.com/jasonwillschiu/code4context-com"));
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

      let skipBuild = false;

      if (shouldBuild) {
        // First check if any .go files have changed
        const hasGoChanges = await hasGoFileChanges();

        if (!hasGoChanges) {
          console.log(green('♻️  No .go files changed, skipping build entirely'));
          skipBuild = true;
        } else {
          // Check if we need to build by comparing content hash with latest version
          if (shouldUploadR2) {
            const requiredEnvVars = ['R2_ACCESS_KEY_ID', 'R2_SECRET_ACCESS_KEY', 'R2_BUCKET_NAME', 'R2_ENDPOINT'];
            const missingVars = requiredEnvVars.filter(varName => !process.env[varName]);

            if (missingVars.length === 0) {
              const awsEnv = {
                ...process.env,
                AWS_ACCESS_KEY_ID: process.env.R2_ACCESS_KEY_ID,
                AWS_SECRET_ACCESS_KEY: process.env.R2_SECRET_ACCESS_KEY,
                AWS_DEFAULT_REGION: 'auto'
              };

              const bucket = process.env.R2_BUCKET_NAME;
              const endpoint = process.env.R2_ENDPOINT;

              console.log(cyan('🔍 Checking if build is needed...'));
              const currentHash = await calculateContentHash();
              const latestVersionData = await getLatestVersionMetadata(awsEnv, bucket, endpoint);

              if (latestVersionData && latestVersionData.metadata.content_hash === currentHash) {
                console.log(green(`♻️  Content unchanged (hash: ${currentHash.substring(0, 8)}...), skipping build`));
                skipBuild = true;
              } else {
                await buildCrossPlatform(version);
              }
            } else {
              await buildCrossPlatform(version);
            }
          } else {
            await buildCrossPlatform(version);
          }
        }
      }

      if (shouldUploadR2) {
        await uploadToR2(version, skipBuild, summary, description);
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
  console.log(cyan("  bun run cicd.js --build --commit --tag --release  # Same as --full-release"));
  console.log(cyan("  bun run cicd.js --install-guide                   # Show installation guide for AI tools"));
  process.exit(1);
}
