# Setting Up a GitHub App for Release Please

This document explains how to set up a GitHub App to use with Release Please for automated versioning and releases.

## Why a GitHub App?

GitHub Actions workflows running with the default `GITHUB_TOKEN` have limitations when it comes to creating pull requests that trigger other workflows. Using a GitHub App provides:

1. Higher rate limits than personal access tokens
2. Fine-grained permissions
3. The ability to trigger workflows from PRs created by the app
4. Better security than using personal access tokens

## Step-by-Step Setup

### 1. Create a GitHub App

1. Go to your GitHub account settings
2. Navigate to "Developer settings" > "GitHub Apps" > "New GitHub App"
3. Fill in the required information:
   - **Name**: `Release Please App` (or any name you prefer)
   - **Homepage URL**: Your repository URL
   - **Webhook**: Uncheck "Active" (we don't need webhooks)
   - **Permissions**:
     - Repository permissions:
       - Contents: Read & write
       - Pull requests: Read & write
       - Metadata: Read-only
   - **Where can this GitHub App be installed?**: Only on this account

4. Click "Create GitHub App"

### 2. Generate a Private Key

1. After creating the app, navigate to its settings page
2. Scroll down to the "Private keys" section
3. Click "Generate a private key"
4. Save the downloaded `.pem` file securely

### 3. Install the App

1. On the app's settings page, click "Install App"
2. Choose the repository where you want to install the app
3. Click "Install"

### 4. Note the App ID

1. On the app's settings page, note the "App ID" at the top of the page

### 5. Add Secrets to Your Repository

1. Go to your repository settings
2. Navigate to "Secrets and variables" > "Actions"
3. Add the following secrets:
   - `GH_APP_ID`: The App ID you noted earlier
   - `GH_APP_PRIVATE_KEY`: The entire contents of the `.pem` file you downloaded

## Using the GitHub App in Workflows

The Release Please workflow is already configured to use these secrets with the `tibdex/github-app-token` action to generate a token:

```yaml
- name: Generate token
  id: generate_token
  uses: tibdex/github-app-token@v2
  with:
    app_id: ${{ secrets.GH_APP_ID }}
    private_key: ${{ secrets.GH_APP_PRIVATE_KEY }}

- name: Run Release Please
  uses: googleapis/release-please-action@v4
  with:
    token: ${{ steps.generate_token.outputs.token }}
    # other configuration...
```

## Troubleshooting

If you encounter issues with the GitHub App:

1. **Check Permissions**: Ensure the app has the necessary repository permissions
2. **Verify Installation**: Make sure the app is installed on the repository
3. **Check Secrets**: Ensure the secrets are correctly set in the repository
4. **Review Logs**: Check the workflow logs for any error messages

For more information, see the [GitHub Apps documentation](https://docs.github.com/en/developers/apps/getting-started-with-apps/about-apps).
