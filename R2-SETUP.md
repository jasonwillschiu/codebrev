# Cloudflare R2 Setup Guide

This guide explains how to set up Cloudflare R2 for hosting your code4context binaries.

## 🔑 Required API Keys and Setup

### 1. Create Cloudflare R2 Bucket

1. **Log into Cloudflare Dashboard**
   - Go to [dash.cloudflare.com](https://dash.cloudflare.com)
   - Navigate to **R2 Object Storage**

2. **Create a Bucket**
   ```
   Bucket Name: code4context-binaries (or your preferred name)
   Location: Auto (or choose specific region)
   ```

3. **Configure Public Access (Optional)**
   - Go to **Settings** → **Public Access**
   - Enable **Custom Domains** if you want a custom URL
   - Or use the default R2.dev subdomain

### 2. Generate API Tokens

1. **Go to API Tokens**
   - Dashboard → **My Profile** → **API Tokens**
   - Click **Create Token**

2. **Create R2 Token**
   ```
   Token Name: code4context-r2-upload
   Permissions:
   - Zone:Zone Settings:Read (if using custom domain)
   - Zone:Zone:Read (if using custom domain)  
   - Account:Cloudflare R2:Edit
   
   Account Resources: Include - Your Account
   Zone Resources: Include - Specific zone (if using custom domain)
   ```

3. **Get Account ID**
   - Found in the right sidebar of your Cloudflare dashboard
   - Copy the **Account ID**

### 3. Environment Variables Setup

Create a `.env` file or set these environment variables:

```bash
# Required for R2 uploads
export R2_ACCESS_KEY_ID="your_r2_access_key_id"
export R2_SECRET_ACCESS_KEY="your_r2_secret_access_key"
export R2_BUCKET_NAME="code4context-binaries"
export R2_ENDPOINT="https://your-account-id.r2.cloudflarestorage.com"

# Optional: Public URL for downloads
export R2_PUBLIC_URL="https://your-custom-domain.com"
# OR use the default R2.dev URL:
# export R2_PUBLIC_URL="https://pub-your-bucket-id.r2.dev"
```

### 4. AWS CLI Installation

The R2 upload uses AWS CLI with S3-compatible commands:

```bash
# macOS
brew install awscli

# Ubuntu/Debian
sudo apt install awscli

# Windows
# Download from: https://aws.amazon.com/cli/
```

## 🚀 Usage Examples

### Build and Upload to R2

```bash
# Complete R2 release workflow
bun run cicd.js --r2-release

# Individual steps
bun run cicd.js --build
bun run cicd.js --upload-r2

# With git operations
bun run cicd.js --build --commit --tag --upload-r2
```

### Install from R2

```bash
# Set R2 as source
export USE_R2=true
export R2_BASE_URL="https://your-domain.com"

# Install specific version
curl -fsSL https://raw.githubusercontent.com/jasonwillschiu/code4context-com/main/install.sh | sh -s -- --version 0.1.2

# Or use flags directly
curl -fsSL https://raw.githubusercontent.com/jasonwillschiu/code4context-com/main/install.sh | sh -s -- --use-r2 --r2-url https://your-domain.com --version 0.1.2
```

## 📁 R2 File Structure

Your R2 bucket will be organized as:

```
code4context-binaries/
├── latest-version.txt                    # Contains latest version number
└── releases/
    ├── v0.1.0/
    │   ├── code4context-darwin-amd64
    │   ├── code4context-darwin-arm64
    │   ├── code4context-linux-amd64
    │   ├── code4context-linux-arm64
    │   ├── code4context-windows-amd64.exe
    │   └── code4context-windows-arm64.exe
    ├── v0.1.1/
    │   └── ... (same structure)
    └── v0.1.2/
        └── ... (same structure)
```

## 🔧 Troubleshooting

### Common Issues

1. **"Missing required environment variables"**
   ```bash
   # Check all required vars are set
   echo $R2_ACCESS_KEY_ID
   echo $R2_SECRET_ACCESS_KEY
   echo $R2_BUCKET_NAME
   echo $R2_ENDPOINT
   ```

2. **"AWS CLI not found"**
   ```bash
   # Install AWS CLI
   brew install awscli  # macOS
   # or follow installation guide for your OS
   ```

3. **"Access Denied" errors**
   - Verify your API token has R2:Edit permissions
   - Check the bucket name matches exactly
   - Ensure the endpoint URL is correct

4. **"Bucket not found"**
   - Verify bucket exists in your Cloudflare dashboard
   - Check the bucket name in your environment variables
   - Ensure you're using the correct account

### Debug Mode

Enable verbose AWS CLI output:

```bash
# Add --debug flag to see detailed S3 operations
export AWS_CLI_DEBUG=1
bun run cicd.js --upload-r2
```

## 💰 Cost Considerations

### R2 Pricing (as of 2024)
- **Storage**: $0.015/GB/month
- **Class A Operations** (PUT, COPY, POST, LIST): $4.50/million
- **Class B Operations** (GET, SELECT): $0.36/million
- **Egress**: Free up to 10GB/month per account

### Example Costs for code4context
- **6 binaries × 10MB each = 60MB per release**
- **Storage**: ~$0.001/month per release
- **Uploads**: ~$0.000027 per release (6 PUT operations)
- **Downloads**: Free for first 10GB/month

**Extremely cost-effective** compared to other hosting solutions!

## 🔒 Security Best Practices

1. **Use Scoped API Tokens**
   - Only grant R2:Edit permissions
   - Limit to specific account/bucket if possible

2. **Rotate Keys Regularly**
   - Generate new API tokens every 90 days
   - Update environment variables

3. **Environment Variables**
   - Never commit API keys to git
   - Use `.env` files (add to `.gitignore`)
   - Use CI/CD secrets for automation

4. **Public Access**
   - Only enable if you need public downloads
   - Consider using signed URLs for private access

## 🌐 Custom Domain Setup (Optional)

1. **Add Custom Domain in R2**
   - Go to your bucket → **Settings** → **Custom Domains**
   - Add your domain (e.g., `cdn.yoursite.com`)

2. **DNS Configuration**
   - Add CNAME record: `cdn.yoursite.com` → `your-bucket.r2.cloudflarestorage.com`

3. **Update Environment Variable**
   ```bash
   export R2_PUBLIC_URL="https://cdn.yoursite.com"
   ```

This gives you a branded download URL for your binaries!