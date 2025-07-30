# GitHub Secrets 配置指南

为了使GitHub Actions工作流程正常运行，需要在GitHub仓库设置中配置以下secrets：

## 必需的Secrets

### Apple代码签名和公证 (仅macOS版本需要)

这些secrets仅在发布标签时（如 `v1.0.0`）的macOS代码签名和公证过程中使用：

1. **APPLE_CERTIFICATE_BASE64** - Apple Developer证书
   - 导出Developer ID Application证书为.p12文件
   - 使用base64编码：`base64 -i certificate.p12 | pbcopy`

2. **APPLE_CERTIFICATE_PASSWORD** - 证书密码
   - 导出.p12文件时设置的密码

3. **APPLE_DEVELOPER_ID** - 开发者身份
   - 格式：`Your Name (Team ID)`
   - 示例：`John Doe (ABC123DEF4)`

4. **APPLE_ID** - Apple ID邮箱
   - 用于公证的Apple ID账号

5. **APPLE_APP_SPECIFIC_PASSWORD** - App专用密码
   - 在Apple ID管理页面生成的专用密码
   - 不是Apple ID的登录密码

6. **APPLE_TEAM_ID** - 团队ID
   - 10位字符的团队标识符
   - 可在Apple Developer账户中找到

## 配置步骤

1. 进入GitHub仓库页面
2. 点击 **Settings** 标签
3. 在左侧菜单中选择 **Secrets and variables** → **Actions**
4. 点击 **New repository secret**
5. 分别添加上述secrets

## 注意事项

- 这些secrets仅用于tagged releases（如 `v1.0.0`）
- 如果没有配置这些secrets，workflow仍会构建所有平台的二进制文件，但macOS版本不会被签名和公证
- Pull requests和普通push不会触发签名流程，因此不需要这些secrets即可测试构建过程

## 获取Apple开发者证书

1. 在Apple Developer网站申请Developer ID Application证书
2. 下载并安装到macOS Keychain
3. 导出为.p12格式：
   ```bash
   # 在macOS上导出证书
   security find-identity -v -p codesigning
   # 找到你的Developer ID，然后导出
   ```

## 测试workflow

可以先创建一个测试标签来验证workflow：

```bash
git tag v0.0.1-test
git push origin v0.0.1-test
```

检查GitHub Actions页面确认构建是否成功。

## 🔐 获取Apple开发者证书

### 1. 在Keychain Access中导出证书

1. 打开 **Keychain Access** (钥匙串访问)
2. 在左侧选择 **login** 钥匙串
3. 在分类中选择 **My Certificates** (我的证书)
4. 找到您的 **Developer ID Application** 证书
5. 右键点击证书 → **Export "Developer ID Application: ..."**
6. 选择文件格式为 **Personal Information Exchange (.p12)**
7. 设置一个强密码并记住它 (这将是 `APPLE_CERTIFICATE_PASSWORD`)

### 2. 转换为Base64

在终端中运行以下命令：

```bash
# 将证书转换为base64编码
base64 -i /path/to/your/certificate.p12 -o certificate.base64

# 复制base64内容
cat certificate.base64
```

将输出的base64字符串复制并设置为 `APPLE_CERTIFICATE_BASE64` secret。

## 🔑 获取App专用密码

1. 访问 [Apple ID 账户页面](https://appleid.apple.com/)
2. 登录您的Apple ID
3. 在 **Security** (安全性) 部分找到 **App-Specific Passwords** (App专用密码)
4. 点击 **Generate Password** (生成密码)
5. 输入密码标签 (例如: "GitHub Actions Notarization")
6. 复制生成的密码，设置为 `APPLE_APP_SPECIFIC_PASSWORD` secret

## 🆔 获取团队ID和开发者ID

### 获取团队ID
1. 访问 [Apple Developer Portal](https://developer.apple.com/account/)
2. 登录后，在右上角可以看到您的团队ID (10位字符)

### 获取开发者ID
在终端中运行：
```bash
# 查看所有可用的代码签名身份
security find-identity -v -p codesigning

# 找到类似这样的输出：
# "Developer ID Application: Your Name (XXXXXXXXXX)"
```

## 🛠️ 测试配置

配置完成后，您可以：

1. 创建一个新的Git标签来触发工作流：
   ```bash
   git tag v1.0.0
   git push origin v1.0.0
   ```

2. 或者手动触发工作流：
   - 在GitHub仓库页面进入 `Actions` 标签
   - 选择 `Build and Notarize` 工作流
   - 点击 `Run workflow`

## 🚨 安全注意事项

1. **永远不要**将证书或密码直接写在代码或配置文件中
2. 定期更新App专用密码
3. 监控GitHub Actions的使用情况
4. 如果怀疑secrets泄露，立即重新生成所有相关密码和证书

## 📚 相关链接

- [Apple Developer Documentation - Notarizing macOS Software](https://developer.apple.com/documentation/security/notarizing_macos_software_before_distribution)
- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Apple ID App-Specific Passwords](https://support.apple.com/en-us/HT204397)

## ❓ 故障排除

### 公证失败常见原因：
1. **证书过期**：检查Developer ID证书是否仍然有效
2. **App专用密码错误**：重新生成App专用密码
3. **团队ID不匹配**：确认使用正确的10位团队标识符
4. **二进制文件未签名**：确认codesign步骤成功执行

### 查看详细错误信息：
在GitHub Actions日志中查看具体的错误信息，通常包含解决问题的线索。
