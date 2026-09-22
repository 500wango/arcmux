/*
Copyright (C) 2023-2026 ArcMux contributors

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.
*/
export function getDefaultTermsAndConditions(
  siteName: string,
  lang = 'en'
): string {
  const normalizedLang = lang.toLowerCase()
  const resolvedName = siteName?.trim() || 'ArcMux'

  if (
    normalizedLang.startsWith('zh-tw') ||
    normalizedLang.startsWith('zh-hk')
  ) {
    return `# 服務條款 (Terms & Conditions)

*最近更新日期：2026年9月*

歡迎使用 **${resolvedName}**。本服務條款（下稱「本條款」或「本協議」）規範您對我們網站、應用程式介面（API）、多路中繼閘道、管理控制台及相關開發者工具（統稱「本服務」）的存取與使用。

當您建立帳戶、生成 API 金鑰、調用介面或使用本服務的任何部分時，即表明您已閱讀、理解並完全同意受本條款的約束。如果您不同意本條款的任何內容，請勿使用本服務。

---

### 1. 服務描述與平台定位

${resolvedName} 是一個高並發、彈性的 AI API 聚合與智慧路由閘道。本服務為開發者提供統一標準通訊協定，將請求智慧路由、負載均衡至各大上游人工智慧大模型供應商（包括但不限於 OpenAI、Anthropic、Google、DeepSeek 等）。

您知悉並確認：
- ${resolvedName} 作為技術中繼與路由代理基礎設施運行。除非另有明確聲明，本平台不訓練、不持有亦不直接擁有各第三方上游大模型的底層權重及知識庫。
- 各模型的可用性、回應延遲、上下文長度及生成表現受對應上游供應商營運狀況約束。

---

### 2. 使用者帳戶與 API 金鑰規範

1. **註冊資格**：您須年滿 18 週歲（或具備您所在司法管轄區規定的完全民事行為能力），方可註冊帳戶並使用本服務。
2. **帳戶安全**：您應妥善保管帳戶登入憑證、密碼及生成的 API Key。透過您的帳戶或 API Key 發起的所有操作、調用消耗及法律後果，均由您自行承擔。
3. **金鑰洩漏處置**：如發現任何未經授權的存取、憑據洩漏或安全異常，您應立即在控制台中撤銷受損金鑰，並聯繫管理員協助處置。

---

### 3. 可接受使用政策 (AUP)

您承諾嚴格按照法律法規及本條款合法合規使用本服務，**嚴禁**從事以下行為：

- **違法違規活動**：利用本服務生成、傳播含有違法犯罪、恐怖主義、極端暴力、色情低俗、仇恨言論或侵犯他人合法權益的內容。
- **系統破壞與惡意攻擊**：對本平台伺服器、閘道、資料庫進行掃描、滲透測試、DoS/DDoS 阻斷服務攻擊、暴力破解濫用或任何破壞基礎設施穩定性的行為。
- **規避管控與作弊**：試圖規避速率限制（Rate Limit）、並發限制、計費計量、認證鑑權或存取控制機制。
- **侵犯第三方權益**：侵犯任何自然人或法人的智慧財產權、隱私權、商業秘密等合法權益。
- **違反上游政策**：故意利用本服務規避上游大模型服務商（如 OpenAI、Anthropic、Google 等）的使用政策與安全規範。

---

### 4. 額度計費、儲值與支付政策

1. **計費規則**：本服務按照 Token 消耗量、請求參數、模型倍率以及價格頁面公示的費率標準即時扣減帳戶額度。
2. **預付費與退款說明**：帳戶儲值與額度兌換屬於數位技術資源預付費。由於 AI 計算成本具有即時消耗、不可逆轉的特性，**所有儲值額度一旦生效或消耗，均不支援退款**，法律法規另有強制性規定的除外。
3. **速率與並發控制**：為維護整體服務品質，平台有權根據運行狀況設定或調整各使用者、分組的請求頻率限制（QPS/RPM/TPM）及並發連線數。
4. **價格與倍率調整**：上游供應商價格調整或匯率變動時，平台有權調整模型倍率與價格，更新將及時在價格與模型頁面公佈。

---

### 5. 服務可用性與免責聲明

1. **現狀提供**：本服務按「現狀」（AS IS）與「可獲得性」（AS AVAILABLE）基礎提供。在適用法律允許的最大範圍內，平台不提供任何明示或暗示的保證，包括但不限於適銷性、特定用途適用性及不侵權保證。
2. **不可抗力與中斷**：平台雖配備自動故障轉移與健康探測機制，但無法保證服務絕對不發生中斷、延遲或錯誤。因上游服務商故障、主幹網路波動、駭客攻擊或不可抗力導致的服務波動，平台不承擔賠償責任。
3. **AI 生成內容免責**：平台對上游模型輸出內容的真實性、準確性、完整性與安全性不做擔保，使用者應對依賴或應用該內容所產生的決策及成果承擔全部責任。

---

### 6. 智慧財產權與內容所有權

1. **平台權益**：${resolvedName} 的軟體系統、架構設計、前端介面、商標標識及文件資料均歸屬於本平台所有者或開源授權方。
2. **輸入與輸出所有權**：在您與平台之間，您保留對提交的輸入資料（Prompt）的合法權益；模型生成結果（Output）的智慧財產權歸屬依適用法律及對應上游服務商協議確定。

---

### 7. 責任限制

在法律允許的最大限度內，${resolvedName} 及其營運方、開發者對任何因使用或無法使用本服務而引起的間接、附帶、懲罰性、特殊或後果性損害（包括利潤損失、資料遺失、業務中斷等）概不負責。

---

### 8. 違規處罰與服務終止

若您違反本條款或實施損害平台及其他使用者權益的行為，平台有權視違規情節單方面採取限制並發、凍結額度、暫停 API 存取或永久註銷帳戶等措施，且不予補償或退款。

---

### 9. 條款修訂

平台有權依據法律法規變化或業務營運需要適時修訂本條款，修訂後的條款一經發布即行生效。若您繼續使用本服務，即視為接受修訂後的條款。

---

### 10. 聯絡方式

如對本服務條款有任何疑問、意見或回饋，請透過平台官方支援管道、公告欄或 GitHub 社群與我們取得聯絡。`
  }

  if (normalizedLang.startsWith('zh')) {
    return `# 服务条款 (Terms & Conditions)

*最近更新日期：2026年9月*

欢迎使用 **${resolvedName}**。本服务条款（下称“本条款”或“本协议”）规范您对我们网站、应用程序编程接口（API）、多路中继网关、管理控制台及相关开发者工具（统称“本服务”）的访问与使用。

当您创建账户、生成 API 密钥、调用接口或使用本服务的任何部分时，即表明您已阅读、理解并完全同意受本条款的约束。如果您不同意本条款的任何内容，请勿使用本服务。

---

### 1. 服务描述与平台定位

${resolvedName} 是一个高并发、弹性的 AI API 聚合与智能分发网关。本服务为开发者提供统一标准协议，将请求智能路由、负载均衡至各大上游人工智能大模型供应商（包括但不限于 OpenAI、Anthropic、Google、DeepSeek 等）。

您知悉并确认：
- ${resolvedName} 作为技术中继与路由代理基础设施运行。除非另有明确声明，本平台不训练、不持有亦不直接拥有各第三方上游大模型的底层权重及知识库。
- 各模型的可用性、响应延迟、上下文长度及生成表现受对应上游供应商运营状况约束。

---

### 2. 用户账户与 API 密钥规范

1. **注册资格**：您须年满 18 周岁（或具备您所在司法管辖区规定的完全民事行为能力），方可注册账户并使用本服务。
2. **账户安全**：您应妥善保管账户登录凭证、密码及生成的 API Key。通过您的账户或 API Key 发起的所有操作、调用消耗及法律后果，均由您自行承担。
3. **密钥泄露处理**：如发现任何未经授权的访问、凭据泄露或安全异常，您应立即在控制台中吊销受损密钥，并联系管理员协助处置。

---

### 3. 可接受使用政策 (AUP)

您承诺严格按照法律法规及本条款合法合规使用本服务，**严禁**从事以下行为：

- **违法违规活动**：利用本服务生成、传播含有违法犯罪、恐怖主义、极端暴力、色情低俗、仇恨言论或侵犯他人合法权益的内容。
- **系统破坏与恶意攻击**：对本平台服务器、网关、数据库进行扫描、渗透测试、DoS/DDoS 拒绝服务攻击、爆破滥用或任何破坏基础设施稳定性的行为。
- **规避管控与作弊**：试图规避速率限制（Rate Limit）、并发限制、计费计量、认证鉴权或访问控制机制。
- **侵犯第三方权益**：侵犯任何自然人或法人的知识产权、隐私权、商业秘密等合法权益。
- **违反上游政策**：故意利用本服务规避上游大模型服务商（如 OpenAI、Anthropic、Google 等）的使用政策与安全规范。

---

### 4. 额度计费、充值与支付政策

1. **计费规则**：本服务按照 Token 消耗量、请求参数、模型倍率以及价格页面公示的费率标准实时扣减账户额度。
2. **预付费与退款说明**：账户充值与额度兑换属于数字技术资源预付费。由于 AI 计算成本具有即时消耗、不可逆转的特性，**所有充值额度一旦生效或消耗，均不支持退款**，法律法规另有强制性规定的除外。
3. **速率与并发控制**：为维护整体服务质量，平台有权根据运行状况设定或调整各用户、分组的请求频率限制（QPS/RPM/TPM）及并发通道数。
4. **价格与倍率调整**：上游供应商价格调整或汇率变动时，平台有权调整模型倍率与价格，更新将及时在价格与模型页面公布。

---

### 5. 服务可用性与免责声明

1. **现状提供**：本服务按“现状”（AS IS）与“可获得性”（AS AVAILABLE）基础提供。在适用法律允许的最大范围内，平台不提供任何明示或暗示的保证，包括但不限于适销性、特定用途适用性及不侵权保证。
2. **不可抗力与中断**：平台虽配备自动故障转移与健康探测机制，但无法保证服务绝对不发生中断、延迟或错误。因上游服务商故障、主干网络波动、黑客攻击或不可抗力导致的服务波动，平台不承担赔偿责任。
3. **AI 生成内容免责**：平台对上游模型输出内容的真实性、准确性、完整性与安全性不做担保，用户应对依赖或应用该内容所产生的决策及成果承担全部责任。

---

### 6. 知识产权与内容所有权

1. **平台权益**：${resolvedName} 的软件系统、架构设计、前端界面、商标标识及文档资料均归属于本平台所有者或开源授权方。
2. **输入与输出所有权**：在您与平台之间，您保留对提交的输入数据（Prompt）的合法权益；模型生成结果（Output）的知识产权归属依适用法律及对应上游服务商协议确定。

---

### 7. 责任限制

在法律允许的最大限度内，${resolvedName} 及其运营方、开发者对任何因使用或无法使用本服务而引起的间接、附带、惩罚性、特殊或后果性损害（包括利润损失、数据丢失、业务中断等）概不负责。

---

### 8. 违规处罚与服务终止

若您违反本条款或实施损害平台及其他用户权益的行为，平台有权视违规情节单方面采取限制并发、冻结额度、暂停 API 访问或永久注销账户等措施，且不予补偿或退款。

---

### 9. 条款修订

平台有权依据法律法规变化或业务运营需要适时修订本条款，修订后的条款一经发布即行生效。若您继续使用本服务，即视为接受修订后的条款。

---

### 10. 联系方式

如对本服务条款有任何疑问、意见或反馈，请通过平台官方支持渠道、公告栏或 GitHub 社区与我们取得联系。`
  }

  return `# Terms & Conditions

*Last updated: September 2026*

Welcome to **${resolvedName}**. These Terms & Conditions ("Terms", "Agreement") govern your access to and use of our website, application programming interfaces (APIs), proxy multiplexing gateway, management dashboard, and associated developer tools (collectively, the "Service").

By creating an account, generating an API key, accessing our endpoints, or using any part of the Service, you acknowledge that you have read, understood, and agree to be bound by these Terms. If you do not agree, you must not access or use the Service.

---

### 1. Description of Service

${resolvedName} provides a high-throughput, unified AI API gateway and proxy multiplexer. Our service routes, load-balances, and monitors API requests from clients to various upstream artificial intelligence model providers (including but not limited to OpenAI, Anthropic, Google, DeepSeek, and other supported providers).

You acknowledge and agree that:
- ${resolvedName} functions as an intermediary routing and proxy infrastructure. We do not develop, train, or own the underlying artificial intelligence models provided by third-party upstream platforms unless explicitly specified.
- The availability, output quality, latency, context windows, and functional capabilities of individual models depend on the respective upstream providers.

---

### 2. User Accounts & API Credentials

1. **Eligibility**: You must be at least 18 years old (or the legal age of majority in your jurisdiction) and capable of entering into legally binding contracts to create an account and use the Service.
2. **Account Security**: You are solely responsible for maintaining the confidentiality of your account credentials, passwords, and API keys. Any action, request, token consumption, or liability incurred under your account or using your API keys is considered your own act.
3. **Key Management**: You agree to immediately revoke any compromised API keys via the dashboard and notify us if you detect unauthorized access or security incidents related to your account.

---

### 3. Acceptable Use Policy (AUP)

You agree to use ${resolvedName} strictly for lawful purposes and in full compliance with these Terms, applicable laws, and upstream provider usage policies. You must **NOT**:

- **Illegal Activities**: Use the Service to generate, transmit, or process content related to illegal acts, child sexual abuse material (CSAM), terrorism, violence, hate speech, or harassment.
- **System Abuse & Attacks**: Attempt to disrupt, compromise, or overload our servers, networks, load balancers, or database instances through denial-of-service (DoS/DDoS) attacks, brute-force requests, or vulnerability exploitation.
- **Bypassing Protections**: Interfere with or circumvent authentication, rate limits, quota accounting, billing calculations, or access control mechanisms.
- **Infringement**: Infringe upon intellectual property, privacy, publicity, or other legal rights of any third party.
- **Upstream Policy Violation**: Utilize the Service in a manner that intentionally evades or violates the Terms of Service, Acceptable Use Policies, or safety restrictions established by the upstream AI providers (e.g., OpenAI, Anthropic, Google).

---

### 4. Quotas, Billing, & Payments

1. **Billing Mechanics**: Access to AI models through ${resolvedName} is billed according to token usage, request parameters, model ratios, and pricing schedules displayed on our Pricing page. Quota is deducted in real-time or near real-time upon request execution.
2. **Pre-paid Quotas & Top-ups**: Purchased quotas, credits, and balance top-ups represent advance payment for computing and proxy resources. Due to the immediate delivery of digital resources and irreversible token consumption with upstream providers, **all payments and top-ups are non-refundable**, except where required by mandatory applicable law.
3. **Rate Limiting & Concurrency**: We reserve the right to establish and adjust rate limits (requests per minute, tokens per minute, and concurrent request limits) to prevent abuse and protect overall infrastructure stability.
4. **Price Adjustments**: Upstream providers may update their pricing models at any time. ${resolvedName} reserves the right to modify model pricing, billing ratios, and exchange rates with reasonable notice published on the platform.

---

### 5. Service Availability & Disclaimers

1. **"As-Is" Provision**: The Service is provided on an **"AS IS"** and **"AS AVAILABLE"** basis. To the maximum extent permitted by law, ${resolvedName} disclaims all warranties of any kind, whether express, implied, statutory, or otherwise, including implied warranties of merchantability, fitness for a particular purpose, title, and non-infringement.
2. **No Uptime Guarantee**: While we strive for high uptime and resilience through automated health probes and failover routing, we do not warrant that the Service will be uninterrupted, error-free, completely secure, or free from latency spikes caused by upstream provider outages or internet transit disruptions.
3. **AI Output Disclaimer**: We make no warranties regarding the accuracy, truthfulness, completeness, reliability, or safety of any output, text, code, or media generated by upstream AI models. You assume full responsibility for reviewing and verifying AI-generated content before relying upon or deploying it.

---

### 6. Intellectual Property

1. **Platform Rights**: All trademarks, logos, service marks, user interfaces, documentation, code, and platform features of ${resolvedName} are and remain the exclusive property of ${resolvedName} and its licensors.
2. **User Prompts & Outputs**: As between you and ${resolvedName}, you retain whatever intellectual property rights you hold in the prompts and input data you submit. Ownership of AI-generated outputs is governed by the respective terms of the upstream model provider utilized for generation.

---

### 7. Limitation of Liability

To the maximum extent permitted by applicable law, in no event shall ${resolvedName}, its contributors, affiliates, directors, employees, or licensors be liable for any indirect, incidental, special, consequential, or punitive damages—including, without limitation, loss of profits, revenue, data, goodwill, service interruption, computer damage, or system failure—arising out of or in connection with these Terms or your use of or inability to use the Service, whether based on warranty, contract, tort (including negligence), product liability, or any other legal theory.

---

### 8. Suspension & Termination

We reserve the right to suspend, throttle, or terminate your account and API access immediately, without prior notice or liability, in the event of:
- A material breach of these Terms or the Acceptable Use Policy.
- Suspicious, fraudulent, abusive, or harmful activity affecting our platform or other users.
- Compliance with a subpoena, court order, or lawful request from law enforcement or regulatory authorities.

Upon termination, your right to access the Service will immediately cease.

---

### 9. Modifications to Terms

We reserve the right to modify or replace these Terms at our discretion. We will indicate the date of the latest revision at the top of this document. Your continued access to or use of the Service after any revisions become effective constitutes your acceptance of the new terms.

---

### 10. Contact Us

If you have any questions or concerns regarding these Terms & Conditions, please contact us through the official support channels, community forums, or GitHub repository associated with this platform.`
}

export function getDefaultPrivacyPolicy(siteName: string, lang = 'en'): string {
  const normalizedLang = lang.toLowerCase()
  const resolvedName = siteName?.trim() || 'ArcMux'

  if (
    normalizedLang.startsWith('zh-tw') ||
    normalizedLang.startsWith('zh-hk')
  ) {
    return `# 隱私權政策 (Privacy Policy)

*最近更新日期：2026年9月*

**${resolvedName}**（下稱「我們」）深知隱私與個人資料對您的重要性。本隱私權政策旨在向您說明當您造訪我們的網站、註冊使用帳戶或透過 API 調用本平台閘道服務時，我們如何收集、使用、儲存及保護您的資訊，以及您所享有的資料權利。

---

### 1. 我們收集的資訊

為向您提供穩定、安全的 API 聚合與路由轉發服務，我們收集以下必要資訊：

1. **帳戶資訊**：當您註冊帳戶時，我們會收集您的使用者名稱、電子郵件地址及加密密碼。若您使用第三方 OAuth（如 GitHub、LinuxDO、Discord、微信等）快速登入，我們將根據您的授權獲取公開基礎資料識別碼。
2. **API 調用與日誌元資料**：為了實現額度計量、路由尋優和系統防護，閘道會自動記錄：
   - 請求發起時間戳記與存取來源 IP 位址。
   - 調用的模型識別碼、請求路徑與 HTTP 狀態碼。
   - Token 消耗統計（提示 Token、補全 Token、思考 Token）。
   - 上游回應延遲、故障重試狀態及異常錯誤代碼。
3. **交易與支付記錄**：若您購買額度或儲值餘額，第三方支付平台（如 Stripe、Creem、易支付等）會向我們返回訂單編號、交易金額、支付管道及完成狀態。**我們不會在伺服器中儲存您的信用卡號或敏感支付憑據。**
4. **瀏覽器本機偏好**：我們透過 Cookie 及本機儲存記錄您的登入認證憑證、介面主題（明亮/暗黑）及語言偏好設定。

---

### 2. 我們如何使用您的資訊

我們僅將收集的資訊用於以下合法目的：
- **核心服務交付**：驗證使用者身分、鑑權 API Key、執行流量負載均衡與智慧路由，並準確核算與扣減 Token 額度。
- **系統安全與風控**：即時偵測惡意高頻攻擊、暴力破解、DDoS 阻斷服務攻擊及 API 盜刷行為。
- **服務穩定性與品質優化**：統計節點連通率、延遲分佈與故障熔斷指標，持續提升閘道中繼效能。
- **客戶支援與服務通知**：回應您的工單諮詢、協助憑證找回，並在必要時向您發送重要系統變更或安全警報。

---

### 3. 資料流轉與第三方 AI 服務商中繼說明

1. **閘道中繼機制**：當您透過本平台發起模型請求時，您的 Prompt 提示詞與請求參數將透過加密通道即時轉發至提供對應模型的第三方上游服務商（如 OpenAI、Anthropic、Google 等）。
2. **上游服務商政策**：第三方上游服務商將根據其各自的隱私權政策與資料協議處理輸入並返回生成結果。在支援的管道上，我們優先配置企業級不保留資料（Zero Data Retention, ZDR）通道。
3. **閘道儲存原則**：除計費及安全審計所需的元資料（Token 數量、調用時間、耗時等）外，閘道預設以串流（Stream）轉發請求，不對您的請求內容及輸出文字進行留存，除非管理端為特定合規或除錯排查專門啟用了特定日誌策略。

---

### 4. Cookie 與本機儲存

我們僅使用運行所必需的最小化儲存技術：
- **身分認證會話**：透過安全的 Token/Cookie 保持您的登入態。
- **使用者偏好設定**：保存您選擇的介面主題樣式與語言種類。
- 我們**不會**植入任何第三方行為追蹤 Cookie 或廣告推播追蹤組件。

---

### 5. 資料安全與保護措施

我們採取嚴格的技術與組織保障措施確保您的資料安全：
- **通訊加密**：全站強制使用現代化 TLS/HTTPS 協定，保障傳輸通道安全。
- **密碼加密儲存**：所有使用者密碼均採用業界標準高強度雜湊演算法（bcrypt/Argon2）加鹽單向加密，無法被明文還原。
- **存取控制機制**：生產環境資料庫與伺服器實行最小權限原則，部署多重防火牆與安全隔離。

---

### 6. 資訊的共享與揭露

我們絕不會向任何無關第三方出售、出租或交易您的個人資訊。僅在以下必要情況下，我們可能共享或揭露資訊：
- **基礎設施與雲端服務供應商**：經嚴格審查並受保密協議約束的雲端伺服器、資料庫託管及合規支付閘道。
- **法律合規與執法要求**：依據有效法院判令、傳票或法律法規強制要求而必須進行揭露時。
- **重大權益保護**：為調查網路犯罪、制止嚴重侵害平台或其他使用者生命財產安全的情形。

---

### 7. 資料保留與您的權利

1. **保存期限**：您的帳戶資訊在帳戶存續期間持續保留；調用日誌與交易憑證依合規與對帳要求保存合理期限後定期輪轉清理或匿名化處理。
2. **您的權利**：根據適用法律法規，您對自己的個人資訊享有查閱、更正、匯出及請求註銷帳戶的權利。您可以透過個人中心進行管理或聯繫平台支援人員協助辦理。

---

### 8. 隱私權政策的更新

我們可能會根據技術架構演進或法律法規要求適時修訂本隱私權政策。任何重大變更均會在本頁面公佈並更新生效日期。若您繼續使用本服務，即視為同意更新後的政策內容。

---

### 9. 聯絡我們

如您對本隱私權政策有任何疑問、建議或隱私權益行使需求，請透過平台官方管道與我們聯繫。`
  }

  if (normalizedLang.startsWith('zh')) {
    return `# 隐私政策 (Privacy Policy)

*最近更新日期：2026年9月*

**${resolvedName}**（下称“我们”）深知隐私与个人数据对您的重要性。本隐私政策旨在向您说明当您访问我们的网站、注册使用账户或通过 API 调用本平台网关服务时，我们如何收集、使用、存储及保护您的信息，以及您所享有的数据权利。

---

### 1. 我们收集的信息

为向您提供稳定、安全的 API 聚合与路由分发服务，我们收集以下必要信息：

1. **账户信息**：当您注册账户时，我们会收集您的用户名、电子邮箱地址及加密密码。若您使用第三方 OAuth（如 GitHub、LinuxDO、Discord、微信等）快捷登录，我们将根据您的授权获取公开基础资料标识。
2. **API 调用与日志元数据**：为了实现额度计量、路由寻优和系统防护，网关会自动记录：
   - 请求发起时间戳与访问来源 IP 地址。
   - 调用的模型标识、请求路径与 HTTP 状态码。
   - Token 消耗统计（提示 Token、补全 Token、思考 Token）。
   - 上游响应延迟、故障重试状态及异常错误代码。
3. **交易与支付记录**：若您购买额度或充值余额，第三方支付平台（如 Stripe、Creem、易支付等）会向我们返回订单编号、交易金额、支付渠道及完成状态。**我们不会在服务器中存储您的银行卡号或敏感支付凭据。**
4. **浏览器本地偏好**：我们通过 Cookie 及本地存储记录您的登录认证凭证、界面主题（明亮/暗黑）及语言偏好设置。

---

### 2. 我们如何使用您的信息

我们仅将收集的信息用于以下合法目的：
- **核心服务交付**：验证用户身份、鉴权 API Key、执行流量负载均衡与智能路由，并准确核算与扣减 Token 额度。
- **系统安全与风控**：实时侦测恶意高频攻击、暴力破解、DDoS 拒绝服务攻击及 API 盗刷行为。
- **服务稳定性与质量优化**：统计节点连通率、延迟分布与故障熔断指标，持续提升网关中继性能。
- **客户支持与服务通知**：响应您的工单咨询、协助凭据找回，并在必要时向您发送重要系统变更或安全警报。

---

### 3. 数据流转与第三方 AI 服务商中继说明

1. **网关中继机制**：当您通过本平台发起模型请求时，您的 Prompt 提示词与请求参数将通过加密通道实时转发至提供对应模型的第三方上游服务商（如 OpenAI、Anthropic、Google 等）。
2. **上游服务商政策**：第三方上游服务商将根据其各自的隐私政策与数据协议处理输入并返回生成结果。在支持的渠道上，我们优先配置企业级不保留数据（Zero Data Retention, ZDR）通道。
3. **网关存储原则**：除计费及安全审计所需的元数据（Token 数量、调用时间、耗时等）外，网关默认以流式（Stream）转发请求，不对您的请求内容及输出文本进行留存，除非管理端为特定合规或调试排查专门启用了特定日志策略。

---

### 4. Cookie 与本地存储

我们仅使用运行所必需的最小化存储技术：
- **身份认证会话**：通过安全的 Token/Cookie 保持您的登录态。
- **用户偏好设置**：保存您选择的界面主题样式与语言种类。
- 我们**不会**植入任何第三方行为追踪 Cookie 或广告推送跟踪组件。

---

### 5. 数据安全与保护措施

我们采取严格的技术与组织保障措施确保您的数据安全：
- **通信加密**：全站强制使用现代化 TLS/HTTPS 协议，保障传输通道安全。
- **密码加密存储**：所有用户密码均采用行业标准高强度哈希算法（bcrypt/Argon2）加盐单向加密，无法被明文还原。
- **访问控制机制**：生产环境数据库与服务器实行最小权限原则，部署多重防火墙与安全隔离。

---

### 6. 信息的共享与披露

我们绝不会向任何无关第三方出售、出租或交易您的个人信息。仅在以下必要情况下，我们可能共享或披露信息：
- **基础设施与云服务供应商**：经严格审查并受保密协议约束的云服务器、数据库托管及合规支付网关。
- **法律合规与执法要求**：依据有效法院判令、传票或法律法规强制要求而必须进行披露时。
- **重大权益保护**：为调查网络犯罪、制止严重侵害平台或其他用户生命财产安全的情形。

---

### 7. 数据保留与您的权利

1. **保存期限**：您的账户信息在账户存续期间持续保留；调用日志与交易凭证依合规与对账要求保存合理期限后定期轮转清理或匿名化处理。
2. **您的权利**：根据适用法律法规，您对自己的个人信息享有查阅、更正、导出及请求注销账户的权利。您可以通过个人中心进行管理或联系平台支持人员协助办理。

---

### 8. 隐私政策的更新

我们可能会根据技术架构演进或法律法规要求适时修订本隐私政策。任何重大变更均会在本页面公布并更新生效日期。若您继续使用本服务，即视为同意更新后的政策内容。

---

### 9. 联系我们

如您对本隐私政策有任何疑问、建议或隐私权益行使需求，请通过平台官方渠道与我们联系。`
  }

  return `# Privacy Policy

*Last updated: September 2026*

**${resolvedName}** ("we", "us", "our") is dedicated to protecting the privacy and personal data of our users ("you", "your"). This Privacy Policy explains what information we collect, how we process and store it, the purpose of collection, and your rights when you visit our website, register an account, or interface with our AI gateway APIs.

---

### 1. Information We Collect

We collect information necessary to provide, protect, and optimize our API multiplexing services:

1. **Account Credentials**: When registering, you provide an email address, username, and password. If you authenticate via third-party OAuth providers (e.g., GitHub, LinuxDO, Discord, WeChat), we receive public profile identifiers authorized by you.
2. **API Usage & Telemetry Data**: To route requests, calculate token consumption, and safeguard network stability, our systems log:
   - Request timestamps and client IP addresses.
   - Target model identifier, requested endpoint, and HTTP response status code.
   - Token usage metrics (prompt tokens, completion tokens, reasoning tokens).
   - Upstream latency, retry counts, and error diagnostics.
3. **Billing & Payment Records**: If you purchase quota, our payment processing partners (e.g., Stripe, Creem, Epay) provide transaction references, order IDs, amounts, and completion timestamps. **We do not process or store raw payment card numbers on our servers.**
4. **Local Browser Preferences**: We store essential cookies and local browser state to remember your authentication session, theme choice (light/dark mode), and language preferences.

---

### 2. How We Use Your Information

We use collected information solely for the following legitimate purposes:
- **Service Operation**: Authenticating users, routing API requests, enforcing permissions, and deducting token quotas accurately.
- **Security & Fraud Prevention**: Detecting abnormal request spikes, brute-force attempts, DDoS threats, and unauthorized API key sharing.
- **Performance Optimization**: Analyzing routing latencies, error distributions, and channel availability to ensure high uptime.
- **Customer Support & Communications**: Assisting with technical inquiries, account recovery, and delivering critical service announcements.

---

### 3. Data Flow & Upstream AI Relay

1. **Relay Architecture**: When you send an inference request through our gateway, your prompt data and parameters are forwarded to the upstream provider (e.g., OpenAI, Anthropic, Google) hosting the selected model.
2. **Upstream Data Policies**: Upstream AI providers process prompts and generate completions according to their respective privacy policies and enterprise terms. We configure upstream connections to prioritize enterprise non-training zero-data-retention (ZDR) endpoints wherever supported by providers.
3. **Gateway Storage**: By default, our gateway retains metadata (token counts, timestamps, status codes) for quota ledger tracking. Full prompt/completion payloads are processed in-stream without gateway persistence, unless custom administrative logging is explicitly enabled for compliance or debugging.

---

### 4. Cookies & Local Storage

We use minimal, strictly necessary cookies and local storage tokens:
- **Authentication Session**: Secure, HTTP-only or token-based session identifiers to keep you logged in.
- **UI Preferences**: Local storage items for theme selection (light/dark/system) and interface language.
- We do **not** use third-party tracking pixels, cross-site profiling, or behavioral advertising cookies.

---

### 5. Data Security & Protection

We employ rigorous technical and organizational measures to safeguard your information:
- **Encryption**: All communications with our website and API endpoints are protected using modern TLS/HTTPS encryption.
- **Credential Hashing**: User passwords are encrypted using cryptographic hashing algorithms (bcrypt/Argon2) with unique salts.
- **Access Control**: Production databases and administration controls are restricted by role-based permissions and protected behind secure environments.

---

### 6. Information Sharing & Disclosure

We do not sell, rent, or monetize your personal information to third parties. We disclose information only under the following limited conditions:
- **Third-Party Infrastructure**: Trusted hosting, database, Redis, and payment gateway infrastructure providers who operate under contractual data protection obligations.
- **Legal Compliance**: When required to comply with a valid court order, subpoena, or applicable regulatory requirement.
- **Safety & Rights Protection**: When necessary to investigate fraud, enforce our Terms & Conditions, or protect the safety and integrity of our systems and users.

---

### 7. Data Retention & Your Rights

1. **Retention Period**: Account information is retained for the duration of your active account. Usage logs and ledger records are retained for auditing, billing verification, and performance analysis, after which they are periodically purged.
2. **Your Rights**: Depending on your location and applicable privacy regulations (such as GDPR or CCPA), you have the right to:
   - Access and export your account data.
   - Correct inaccurate information in your profile.
   - Delete your account and associated API keys.
3. To exercise your rights, please access your account settings or contact our support team.

---

### 8. Changes to This Privacy Policy

We may periodically update this Privacy Policy to reflect changes in legal requirements or platform architecture. We will post any updates on this page with a revised "Last updated" date. Continued use of our Service after updates confirms your agreement with the revised policy.

---

### 9. Contact Us

If you have questions, inquiries, or privacy-related requests regarding this Privacy Policy, please reach out via platform support or official communication channels.`
}
