# Changelog

## [1.3.0](https://github.com/BernhardRode/aws-infra-sandbox/compare/v1.2.0...v1.3.0) (2025-04-27)


### Features

* remove the policies ([acc8359](https://github.com/BernhardRode/aws-infra-sandbox/commit/acc835972b9347bde9b3760e47e6c2aaa6038dc6))


### Bug Fixes

* refactor the setup script ([0dfd4f2](https://github.com/BernhardRode/aws-infra-sandbox/commit/0dfd4f2b91fdaede7a7c28a1f3cf8379c5777302))

## [1.2.0](https://github.com/BernhardRode/aws-infra-sandbox/compare/v1.1.0...v1.2.0) (2025-04-26)


### Features

* fix deployment to staging/production ([d395f6e](https://github.com/BernhardRode/aws-infra-sandbox/commit/d395f6e51820a30493ddce3d766d649fb2891000))

## [1.1.0](https://github.com/BernhardRode/aws-infra-sandbox/compare/v1.0.0...v1.1.0) (2025-04-26)


### Features

* add DNS setup for ebbo.dev domain ([c6ed8d5](https://github.com/BernhardRode/aws-infra-sandbox/commit/c6ed8d5dde93680c6acaf317aa3d9f145832704f))
* add hosted zone ([1c096e5](https://github.com/BernhardRode/aws-infra-sandbox/commit/1c096e540524421c697bae09b6f17d3160d01c21))
* use only arm64 runners ([90ae376](https://github.com/BernhardRode/aws-infra-sandbox/commit/90ae37620a43a2aa60cced0757bf3b10d771269c))


### Bug Fixes

* close the pr and tear down infra ([8190f95](https://github.com/BernhardRode/aws-infra-sandbox/commit/8190f95c930cd7ec4b6586278a9d815fb3541b76))
* correct TTL variable substitution in DNS setup script ([a466281](https://github.com/BernhardRode/aws-infra-sandbox/commit/a466281ce4eb0c8075f0a42f0bca720dbd3b4c95))
* rename variable ([7c38001](https://github.com/BernhardRode/aws-infra-sandbox/commit/7c380013686aebc6eaa1a9987ae607ddab94af0d))
* split create and delete of pr env ([1b9e6ff](https://github.com/BernhardRode/aws-infra-sandbox/commit/1b9e6fffd2f39a1ced4cf4a459e8c9a59b4e0584))
* update DNS setup script to properly format change batch ([c58b370](https://github.com/BernhardRode/aws-infra-sandbox/commit/c58b3700c60a928f5c8c3949729ad75f5ce627c5))
* update Lambda asset path to use correct build directory ([69a2613](https://github.com/BernhardRode/aws-infra-sandbox/commit/69a2613bdb0d34a938690db639ac0ae2028f5dad))

## 1.0.0 (2025-04-26)


### Features

* add cicd ([1c1823e](https://github.com/BernhardRode/aws-infra-sandbox/commit/1c1823ec899e96f1dd946279b2916fa963f83cda))
* add combined setup command for GitHub Actions and CDK ([826649e](https://github.com/BernhardRode/aws-infra-sandbox/commit/826649e567bfb64cb07a7a3f31ba612ec5abac64))
* add GitHub Actions with AWS IAM Identity Federation setup ([3d61100](https://github.com/BernhardRode/aws-infra-sandbox/commit/3d61100e92c6fe234238afead24fdea581ac2c6f))
* add manual deployment workflow ([6b292f2](https://github.com/BernhardRode/aws-infra-sandbox/commit/6b292f22170af8e4bbb0bc4a736b6952571f9df5))
* add reusable unit test workflow and sequential execution ([663d94c](https://github.com/BernhardRode/aws-infra-sandbox/commit/663d94c3b7217995e45b5338fb9b7da97a364b70))


### Bug Fixes

* add --all flag to CDK commands for multi-stack deployment ([83cb7da](https://github.com/BernhardRode/aws-infra-sandbox/commit/83cb7da8eccf3db8c5a01648d4483225e8f20e58))
* add CDK bootstrap permissions for GitHub Actions roles ([840ebe0](https://github.com/BernhardRode/aws-infra-sandbox/commit/840ebe0696185abf5041398007fae2ed338da6b1))
* cleanup and add staging and production ([c63b31b](https://github.com/BernhardRode/aws-infra-sandbox/commit/c63b31b73e97cbbf8b6683f810e8e401f6377d70))
* ensure IAM roles are always updated with latest permissions ([de141d3](https://github.com/BernhardRode/aws-infra-sandbox/commit/de141d38a69c40d54d80f11e2d38594eddcde854))
* update GitHub Actions workflows and release process ([5331cc7](https://github.com/BernhardRode/aws-infra-sandbox/commit/5331cc7e80ef14c1d468b2b780fbfe5df2c5aa7e))
* update release workflow based on documentation ([194fc16](https://github.com/BernhardRode/aws-infra-sandbox/commit/194fc167838eafd0c2340ff9962c79b19cd037f2))
* update to googleapis/release-please-action ([d1b53b0](https://github.com/BernhardRode/aws-infra-sandbox/commit/d1b53b0883dddcfab0913de5a208990680409879))
