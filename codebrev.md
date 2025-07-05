# Code Structure Outline

This file provides an overview of available functions, types, and variables per file for LLM context.

## File Dependency Graph (LLM Context)

This diagram shows direct file-to-file dependencies to help understand which files are related and may need coordinated changes.

```mermaid
graph TD
    F0["cicd.js"]
    F1["generate_codebrev.go"]
    F2["gitignore/gitignore.go"]
    F3["mermaid/generator.go"]
    F4["outline/dedup.go"]
    F5["outline/types.go"]
    F6["parser/astro.go"]
    F7["parser/go.go"]
    F8["parser/parser.go"]
    F9["writer/writer.go"]
    F10["main.go"]
    F11["FAQSection.astro"]
    F12["FeaturesSection.astro"]
    F13["GoogleAnalytics.astro"]
    F14["Navbar.tsx"]
    F15["OptimizedImage.astro"]
    F16["PricingSection.astro"]
    F17["TestimonialsSection.astro"]
    F18["AppSummary.tsx"]
    F19["PerplexityStyleForm.tsx"]
    F20["ResultsDisplay.tsx"]
    F21["SummaryDisplay.tsx"]
    F22["TranscriptDisplay.tsx"]
    F23["Base.astro"]
    F24["Layout.astro"]
    F25["MarkdownLayout.astro"]
    F26["api.ts"]
    F27["schema.ts"]
    F28["app.astro"]
    F29["index.astro"]
    F30["features.astro"]
    F31["index.astro"]
    F32["pricing.astro"]
    F33["cicd1.js"]

    F1 --> F5
    F1 --> F6
    F1 --> F9
    F3 --> F4
    F6 --> F4
    F7 --> F4
    F8 --> F2
    F8 --> F5
    F9 --> F3
    F9 --> F4
    F10 --> F5
    F10 --> F7
    F10 --> F9
```

## Architecture Overview (Human Context)

This diagram provides a high-level view of the codebase structure with directory groupings and external dependencies.

```mermaid
graph TB
    subgraph internal_gitignore ["internal/gitignore"]
        N0["gitignore/gitignore.go"]
    end

    subgraph internal_mermaid ["internal/mermaid"]
        N1["mermaid/generator.go"]
    end

    subgraph internal_outline ["internal/outline"]
        N2["outline/dedup.go"]
        N3["outline/types.go"]
    end

    subgraph internal_parser ["internal/parser"]
        N4["parser/astro.go"]
        N5["parser/go.go"]
        N6["parser/parser.go"]
    end

    subgraph internal_writer ["internal/writer"]
        N7["writer/writer.go"]
    end

    subgraph root ["root"]
        N8["cicd.js"]
        N9["generate_codebrev.go"]
        N10["main.go"]
    end

    subgraph test-files_frontend4_src_components ["test-files/frontend4/src/components"]
        N11["FAQSection.astro"]
        N12["FeaturesSection.astro"]
        N13["GoogleAnalytics.astro"]
        N14["Navbar.tsx"]
        N15["OptimizedImage.astro"]
        N16["PricingSection.astro"]
        N17["TestimonialsSection.astro"]
    end

    subgraph test-files_frontend4_src_components_app ["test-files/frontend4/src/components/app"]
        N18["AppSummary.tsx"]
        N19["PerplexityStyleForm.tsx"]
        N20["ResultsDisplay.tsx"]
        N21["SummaryDisplay.tsx"]
        N22["TranscriptDisplay.tsx"]
    end

    subgraph test-files_frontend4_src_layouts ["test-files/frontend4/src/layouts"]
        N23["Base.astro"]
        N24["Layout.astro"]
        N25["MarkdownLayout.astro"]
    end

    subgraph test-files_frontend4_src_lib ["test-files/frontend4/src/lib"]
        N26["api.ts"]
        N27["schema.ts"]
    end

    subgraph test-files_frontend4_src_pages ["test-files/frontend4/src/pages"]
        N28["app.astro"]
        N29["features.astro"]
        N30["index.astro"]
        N31["pricing.astro"]
    end

    subgraph test-files_frontend4_src_pages_blog ["test-files/frontend4/src/pages/blog"]
        N32["index.astro"]
    end

    subgraph test-files_onefile ["test-files/onefile"]
        N33["cicd1.js"]
    end

    subgraph external ["External Dependencies"]
        EXT0["@modular-forms/solid"]
        EXT1["~/components/app/AppSu..."]
        EXT2["@solid-primitives/resi..."]
        EXT3["go/ast"]
        EXT4["mark3labs/mcp-go/mcp"]
        EXT5["~icons/carbon/copy-file"]
        EXT6["astro:assets"]
        EXT7["regexp"]
        EXT8["solid-js/web"]
        EXT9["~/components/app/Summa..."]
    end

    N6 --> N0
    N6 --> N3
    N7 --> N1
    N7 --> N3
    N5 --> N2
    N10 --> N3
    N10 --> N5
    N10 --> N7
    N4 --> N3
    N1 --> N2
    N9 --> N3
    N9 --> N4
    N9 --> N7
```

## cicd.js

### Functions
- bold(text)
- buildCrossPlatform(version = null)
- buildLocal()
- calculateContentHash()
- checkBinaryExists(awsEnv, bucket, endpoint, version, platform)
- checkGitStatus()
- checkGitTagExists(version)
- createBunSpinner(initialText = "", opts = {})
- createGitHubRelease(version, summary, description)
- cyan(text)
- findLatestVersionWithBinary(awsEnv, bucket, endpoint, platform, _ignored = 0)
- generateLdflags(version, buildDate, gitCommit)
- getLatestVersionMetadata(awsEnv, bucket, endpoint)
- gitAdd()
- gitCommit(summary, description)
- gitPush()
- gitTag(version, summary)
- green(text)
- hasGoFileChanges()
- mapColor(name)
- parseLatestChangelogEntry()
- red(text)
- render()
- semverToInts(v)
- showInstallationGuide()
- uploadToR2(version, skipBuild = false, releaseSummary = null, releaseDescription = null)
- yellow(text)

### Types
- IMPORTS: bun, path, fs/promises, util

---

## generate_codebrev.go

### Functions
- main()

---

## internal/gitignore/gitignore.go

### Functions
- (Gitignore) ShouldIgnore(path string) -> bool
- (Gitignore) loadGitignoreFile(gitignorePath string)
- (Gitignore) loadGitignoreFromPath(path string)
- (Gitignore) loadGitignoreHierarchy(gitRoot string, scanRoot string)
- (Gitignore) matchPattern(path string, pattern string) -> bool
- New(root string) -> *Gitignore
- findGitRoot(startPath string) -> string

### Types
- Gitignore (methods: ShouldIgnore, loadGitignoreHierarchy, loadGitignoreFile, loadGitignoreFromPath, matchPattern) (fields: Patterns, Root, GitRoot, LoadedDirs)
- Pattern (fields: Pattern, BaseDir)

---

## internal/mermaid/generator.go

### Functions
- GenerateArchitectureOverview(out *outline.Outline) -> string
- GenerateFileDependencyGraph(out *outline.Outline) -> string
- getCleanDepName(dep string) -> string
- getShortFileName(filePath string) -> string
- isLocalImport(imp string) -> bool

---

## internal/outline/dedup.go

### Functions
- (Outline) RemoveDuplicates()

---

## internal/outline/types.go

### Functions
- (Outline) AddDependency(from string, to string)
- (Outline) AddFile(path string) -> *FileInfo
- (Outline) EnsureType(name string) -> *TypeInfo
- New() -> *Outline

### Types
- FileInfo (fields: Path, Functions, Types, Vars, Imports, LocalDeps)
- FunctionInfo (fields: Name, Params, ReturnType)
- Outline (methods: RemoveDuplicates, EnsureType, AddFile, AddDependency) (fields: Files, Types, Vars, Funcs, Dependencies)
- TypeInfo (fields: Fields, Methods)

---

## internal/parser/astro.go

### Functions
- isCustomComponent(tagName string) -> bool
- parseAstroFile(path string, out *outline.Outline, fileInfo *outline.FileInfo) -> error
- parseAstroTemplate(template string, out *outline.Outline, fileInfo *outline.FileInfo)
- parseParameters(paramsStr string) -> []string
- parseTypeScriptContent(content string, out *outline.Outline, fileInfo *outline.FileInfo) -> error
- parseTypeScriptContentRegex(content string, out *outline.Outline, fileInfo *outline.FileInfo) -> error
- parseTypeScriptFile(path string, out *outline.Outline, fileInfo *outline.FileInfo) -> error
- removeDuplicateStrings(slice []string) -> []string
- splitAstroFile(content string) -> string

---

## internal/parser/go.go

### Functions
- extractFunctionInfo(d *ast.FuncDecl) -> outline.FunctionInfo
- parseGoFile(path string, out *outline.Outline, fileInfo *outline.FileInfo, fset *token.FileSet) -> error
- receiverType(expr ast.Expr) -> string
- typeToString(expr ast.Expr) -> string

---

## internal/parser/parser.go

### Functions
- ProcessFiles(root string, out *outline.Outline) -> error
- processFile(path string, info os.FileInfo, out *outline.Outline, fset *token.FileSet) -> error

---

## internal/writer/writer.go

### Functions
- WriteOutlineToFile(out *outline.Outline) -> error
- WriteOutlineToFileWithPath(out *outline.Outline, filePath string) -> error

---

## main.go

### Functions
- addGenerateCodeContextTool(s *server.MCPServer)
- addGetCodeContextTool(s *server.MCPServer)
- generateCodeContext(directoryPath string, outputFile string) -> error
- main()

---

## test-files/frontend4/src/components/FAQSection.astro

### Types
- FAQ (fields: question: string, answer: string)
- faqs: FAQ[]

---

## test-files/frontend4/src/components/FeaturesSection.astro

### Types
- Feature (fields: title: string, description: string, icon: string)
- features: Feature[]

---

## test-files/frontend4/src/components/GoogleAnalytics.astro

### Types
- EXPORTS: Props

---

## test-files/frontend4/src/components/Navbar.tsx

### Functions
- Navbar()
- closeMenu()
- toggleMenu()

### Types
- EXPORTS: Navbar
- IMPORTS: solid-js
- NavLinkItem (fields: href: string, label: string)
- navLinks: NavLinkItem[]

---

## test-files/frontend4/src/components/OptimizedImage.astro

### Types
- COMPONENTS: Picture
- IMPORTS: astro:assets

---

## test-files/frontend4/src/components/PricingSection.astro

### Types
- PricingTier (fields: name: string, price: string, period: string, description: string, features: string[], popular?: boolean, cta: string, ctaHref: string)
- pricingTiers: PricingTier[]

---

## test-files/frontend4/src/components/TestimonialsSection.astro

### Types
- Testimonial (fields: name: string, role: string, company?: string, content: string, avatar: string, rating: number)
- testimonials: Testimonial[]

---

## test-files/frontend4/src/components/app/AppSummary.tsx

### Functions
- AppSummary()
- handleActionComplete(result?: { transcript?: EnhancedTranscriptResponse | null, summary?: string })
- handleActionError(error: string)
- handleActionTrigger(actionType: 'transcript' | 'summarize', triggerSource: 'auto' | 'user' = 'user')

### Types
- AppSummary: Component
- CompletedParams (fields: videoId: string, language: Language, provider?: LLMProvider, model?: LLMModel, actionType: 'transcript' | 'summarize', value: opt.value, label: opt.label, if(import.meta.env.DEBUG), if(isActionInProgress(), if(import.meta.env.DEBUG), setFormError(null), if(isActionInProgress(), if(import.meta.env.DEBUG), if(isActionInProgress(), if(import.meta.env.DEBUG), if(import.meta.env.DEBUG), if(!validationResult), if(import.meta.env.DEBUG), setFormError(validationErrorMsg), if(!providerModelValue), if(import.meta.env.DEBUG), if(providerModelValue === 'auto'), if(parsed && parsed.provider === 'auto' && parsed.model === 'auto'), if(import.meta.env.DEBUG), setIsActionInProgress(false), setIsFetchingTranscriptOnly(false), setIsStartingStream(false), if(parsed && parsed.provider && parsed.provider !== 'auto' && parsed.model !== undefined), setFormError(parseErrorMsg), setIsActionInProgress(false), setIsFetchingTranscriptOnly(false), setIsStartingStream(false), if(!videoId || !language), setFormError(missingParamsMsg), if(currentParams), if(actionType === 'transcript'), if(isSameVideoAndLang && (currentParams.actionType === 'transcript' || currentParams.actionType === 'summarize'), if(import.meta.env.DEBUG), queueMicrotask((), if(transcriptElement), if(import.meta.env.DEBUG), if(isSameVideoAndLang && isSameProvider && isSameModel && currentParams.actionType === 'summarize'), if(import.meta.env.DEBUG), queueMicrotask((), if(summaryElement), if(import.meta.env.DEBUG), if(transcriptElement), if(isSameVideoAndLang && (currentParams.actionType === 'transcript' || currentParams.actionType === 'summarize'), if(import.meta.env.DEBUG), if(import.meta.env.DEBUG), setIsActionInProgress(true), if(actionType === 'transcript'), setIsFetchingTranscriptOnly(true), setIsStartingStream(true), videoId: videoId!, language: language!, provider: provider!, model: model, skipTranscript: skipTranscriptFetch, if(import.meta.env.DEBUG), if(result?.transcript), setTranscriptResult(result.transcript), if(result?.summary != null), setSummaryText(result.summary), if(completedAction), videoId: completedAction.videoId, language: completedAction.language, provider: completedAction.provider, model: completedAction.model, actionType: completedAction.actionType, if(import.meta.env.DEBUG), if(completedActionType === 'transcript'), queueMicrotask((), if(transcriptElement), if(import.meta.env.DEBUG), if(import.meta.env.DEBUG), queueMicrotask((), if(summaryElement), if(import.meta.env.DEBUG), if(transcriptElement), if(import.meta.env.DEBUG), setIsActionInProgress(false), setIsFetchingTranscriptOnly(false), setIsStartingStream(false), setIsStreamingActive(false), setSummaryText(''), setIsActionInProgress(false), setIsFetchingTranscriptOnly(false), setIsStartingStream(false), setIsStreamingActive(false), createEffect((), if(currentVideoId && currentVideoId !== lastVideoId), if(import.meta.env.DEBUG), setLastCompletedParams(null), setActionParams(null), if(import.meta.env.DEBUG), setLastCompletedParams(null), setActionParams(null), formStore: transcriptionForm, Form: Form, Field: Field, providerModelOptions: providerModelOptions, isDisabled: isActionInProgress, onTriggerAction: handleActionTrigger, validateForm: () => validate(transcriptionForm), setFormValue: <T extends keyof TranscriptionFormType>(, fieldName: T, value: Maybe<TranscriptionFormType[T]>, if(value === undefined || value === null), if(fieldName === 'language'), setValue(transcriptionForm, 'language', 'en'), setValue(transcriptionForm, 'providerModel', DEFAULT_PROVIDER_MODEL), setValue(transcriptionForm, '_stream', false), setValue(transcriptionForm, 'url', ''), setValue(transcriptionForm, fieldName, value as any), getFormValue: <T extends keyof TranscriptionFormType>(, fieldName: T, transcriptResult: [transcriptResult, summaryText: [summaryText, chunkSize: [chunkSize)
- IMPORTS: solid-js, solid-js/web, @modular-forms/solid, @kobalte/core/toast, ~icons/carbon/close, ~/lib/api
- SelectOption (fields: value: string, label: string, disabled?: boolean, value: string, label: string, disabled?: boolean)
- perplexityFormProps: PerplexityStyleFormProps
- providerModelOptions: SelectOption[]

---

## test-files/frontend4/src/components/app/PerplexityStyleForm.tsx

### Functions
- PerplexityStyleForm(props)
- copySummary()
- copyTranscript()
- copyTranscriptWithMetadata()
- downloadTranscriptJson()
- findOption(options: SelectOption[], value: string | number | undefined) -> SelectOption | undefined
- formatDisplayDate(dateString: string | undefined) -> string | null
- formatTime(seconds: number) -> string

### Types
- EXPORTS: PerplexityStyleFormProps, PerplexityStyleForm
- FieldName (fields: = keyof T)
- IMPORTS: solid-js, @kobalte/core/text-field, @kobalte/core/toggle-button, @kobalte/core/select, @kobalte/core/button, @kobalte/core/dropdown-menu, @kobalte/core/tooltip, @kobalte/core/toast, @solid-primitives/media, ~icons/carbon/ai-recommend, ~icons/carbon/copy-file, ~icons/carbon/checkmark, ~icons/carbon/close, ~icons/carbon/send, ~icons/carbon/text-wrap, ~/lib/api
- SelectOption (fields: value: string, label: string, disabled?: boolean, value: string, label: string, disabled?: boolean)
- SpecificField (fields: = ReturnType<typeof createForm<TranscriptionFormType>>[1]['Field'])
- SpecificForm (fields: = ReturnType<typeof createForm<TranscriptionFormType>>[1]['Form'])
- chunkSizeOptions: SelectOption[]
- metadataLines: string[]

---

## test-files/frontend4/src/components/app/ResultsDisplay.tsx

### Functions
- ResultsDisplay(props)
- resetInternalState(preserveTranscript: boolean = false)

### Types
- EXPORTS: ResultsDisplayProps, ResultsDisplay
- IMPORTS: solid-js, marked, isomorphic-dompurify, @kobalte/core/toast, @solid-primitives/resize-observer, @solid-primitives/media, ~icons/carbon/close, ~/components/app/TranscriptDisplay, ~/components/app/SummaryDisplay
- metadataUpdate: Partial<SummaryResponse>

---

## test-files/frontend4/src/components/app/SummaryDisplay.tsx

### Functions
- SummaryDisplay(props)

### Types
- EXPORTS: SummaryDisplay
- IMPORTS: solid-js
- SummaryDisplayProps (fields: renderedHtml: string, isStreaming: boolean, isLoading: boolean, metadata: Partial<SummaryResponse> | null, debug?: boolean)

---

## test-files/frontend4/src/components/app/TranscriptDisplay.tsx

### Functions
- TranscriptDisplay(props)
- calculateEndTime(start: number, duration: number) -> number
- formatDisplayDate(dateString: string | undefined) -> string | null
- formatTime(seconds: number) -> string
- scrubLinks(text: string | null | undefined) -> string

### Types
- ChunkedEntry (fields: start: number, duration: number, text: string, entries: TranscriptEntry[])
- EXPORTS: TranscriptDisplay
- IMPORTS: solid-js
- NormalizedEntry (fields: start: number, duration: number, text: string)
- TranscriptDisplayProps (fields: transcript: EnhancedTranscriptResponse | null, chunkSize: () => number)
- chunks: ChunkedEntry[]

---

## test-files/frontend4/src/layouts/Base.astro

### Types
- COMPONENTS: GoogleAnalytics, Navbar
- IMPORTS: ../components/Navbar.tsx, ../components/GoogleAnalytics.astro

---

## test-files/frontend4/src/layouts/Layout.astro

---

## test-files/frontend4/src/layouts/MarkdownLayout.astro

### Types
- COMPONENTS: Base
- EXPORTS: Props
- IMPORTS: ./Base.astro

---

## test-files/frontend4/src/lib/api.ts

### Functions
- extractVideoId(input: string) -> string | null
- fetchData()
- getApiBaseUrl() -> string

### Types
- EXPORTS: getApiBaseUrl, extractVideoId, TranscriptEntry, EnhancedTranscriptResponse, SummaryResponse, TranscriptSummarizeResponse, streamTranscriptAndSummary, FetchEventSourceController, streamSummaryOnly
- IMPORTS: ./schema
- StreamSummaryRequestBody (fields: videoId: string, language: Language, transcript: string, title: string, description: string, provider?: LLMProvider, model?: LLMModel)
- listeners: { [type: string]: ((event: MessageEvent)
- requestBody: StreamSummaryRequestBody

---

## test-files/frontend4/src/lib/schema.ts

### Functions
- getLanguageDisplayName(code: Language) -> string
- getModelDisplayName(code: LLMModel) -> string
- getProviderDisplayName(code: LLMProvider) -> string
- parseProviderModel(value: string | undefined | null)
- validateYoutubeUrl(url: string)

### Types
- EXPORTS: youtubeUrlStringSchema, youtubeUrlSchema, validateYoutubeUrl, languageSchema, Language, languageDisplayNames, getLanguageDisplayName, llmProviderSchema, LLMProvider, providerDisplayNames, getProviderDisplayName, llmModelSchema, LLMModel, modelDisplayNames, getModelDisplayName, ProviderModelOption, allProviderModelOptions, providerModelSchema, ProviderModelValue, parseProviderModel, TranscriptionFormSchema, TranscriptionFormType
- IMPORTS: valibot, ./api

---

## test-files/frontend4/src/pages/app.astro

### Types
- COMPONENTS: Base, AppSummary
- IMPORTS: ~/layouts/Base.astro, ~/components/app/AppSummary.tsx

---

## test-files/frontend4/src/pages/blog/index.astro

### Types
- COMPONENTS: Base
- IMPORTS: ../../layouts/Base.astro

---

## test-files/frontend4/src/pages/features.astro

### Types
- COMPONENTS: Base, FeaturesSection, TestimonialsSection
- IMPORTS: ../layouts/Base.astro, ../components/FeaturesSection.astro, ../components/TestimonialsSection.astro

---

## test-files/frontend4/src/pages/index.astro

### Types
- COMPONENTS: Base
- IMPORTS: ../layouts/Base.astro

---

## test-files/frontend4/src/pages/pricing.astro

### Types
- COMPONENTS: Base, PricingSection, TestimonialsSection, FAQSection
- IMPORTS: ../layouts/Base.astro, ../components/PricingSection.astro, ../components/FAQSection.astro, ../components/TestimonialsSection.astro

---

## test-files/onefile/cicd1.js

### Functions
- bold(text)
- buildCrossPlatform(version = null)
- buildLocal()
- checkGitStatus()
- checkGitTagExists(version)
- createBunSpinner(initialText = "", opts = {})
- createGitHubRelease(version, summary, description)
- cyan(text)
- generateLdflags(version, buildDate, gitCommit)
- gitAdd()
- gitCommit(summary, description)
- gitPush()
- gitTag(version, summary)
- green(text)
- mapColor(name)
- parseLatestChangelogEntry()
- red(text)
- render()
- yellow(text)

### Types
- IMPORTS: bun, path, fs/promises, util

---

