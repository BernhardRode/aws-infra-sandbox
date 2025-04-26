# Manual Deployment

This document explains how to use the manual deployment workflow to deploy specific versions of the application to staging or production environments.

## Overview

The manual deployment workflow allows you to deploy any Git reference (tag, branch, or commit) to either the staging or production environment. This is useful for:

- Deploying hotfixes directly to production
- Rolling back to a previous version
- Testing specific commits in the staging environment
- Deploying release candidates

## Using the Manual Deployment Workflow

### Triggering a Manual Deployment

1. Go to your GitHub repository
2. Click on the "Actions" tab
3. Select the "Manual Deployment" workflow from the list
4. Click the "Run workflow" button
5. Fill in the following parameters:
   - **Environment**: Choose either `staging` or `production`
   - **Git ref**: Enter a tag (e.g., `v1.2.3`), branch name (e.g., `main`), or commit SHA (e.g., `a1b2c3d`)
   - **Version label** (optional): A custom version label for the deployment
6. Click "Run workflow" to start the deployment

### Parameter Details

#### Environment

- **staging**: The pre-production environment for testing
- **production**: The live environment for end users

#### Git ref

You can specify any valid Git reference:
- **Tags**: e.g., `v1.2.3`, `release-2023-04-01`
- **Branches**: e.g., `main`, `feature/new-feature`
- **Commit SHAs**: e.g., `a1b2c3d`, `a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6`

#### Version label

An optional version label for the deployment. If not provided:
- If the Git ref is a version tag (e.g., `v1.2.3`), it will be used as the version
- Otherwise, the short commit SHA will be used

### Deployment Process

When you trigger the manual deployment:

1. The specified Git reference is checked out
2. The code is built and deployed to the selected environment
3. The version label is applied to the deployment
4. A deployment summary is generated with details about the deployment

### Viewing Deployment Status

After triggering the workflow:

1. The workflow will appear in the "Actions" tab
2. You can click on the workflow run to see the deployment progress
3. Once complete, a deployment summary will be available in the workflow run details

## Best Practices

- **Tag releases**: Always tag releases with semantic version numbers (e.g., `v1.2.3`)
- **Test in staging**: Deploy to staging first to verify changes before deploying to production
- **Document deployments**: Add a comment to the workflow run explaining why the manual deployment was needed
- **Coordinate with team**: Inform team members before performing manual deployments to production
- **Verify after deployment**: Always verify that the application is working correctly after a manual deployment
