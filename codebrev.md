# Code Structure Outline

This file provides an overview of available functions, types, and variables per file for LLM context.

## main.go

### Functions
- extractFunctionInfo(d *ast.FuncDecl) -> functionInfo
- extractJSParams(paramStr string) -> []string
- findGitRoot(startPath string) -> string
- isUpperCase(s string) -> bool
- loadGitignore(root string) -> *gitignore
- main()
- parseJSFile(path string, out *outline, fileInfo *fileInfo)
- processFile(path string, info os.FileInfo, out *outline, fset *token.FileSet) -> error
- receiverType(expr ast.Expr) -> string
- removeDuplicates(o *outline)
- typeToString(expr ast.Expr) -> string
- writeOutlineToFile(o *outline)

### Types
- fileInfo (fields: path, functions, types, vars)
- functionInfo (fields: name, params, returnType)
- gitignore (methods: loadGitignoreHierarchy, loadGitignoreFile, loadGitignoreFromPath, shouldIgnore, matchPattern) (fields: patterns, root, gitRoot, loadedDirs)
- gitignorePattern (fields: pattern, baseDir)
- outline (methods: ensureType) (fields: files, types, vars, funcs)
- typeInfo (fields: fields, methods)

### Variables
- currentClass
- dir
- exports
- filePaths
- imports
- result
- returnTypes
- uniqueFileFuncs
- uniqueFileVars
- uniqueFuncs
- uniqueTypes
- uniqueVars

---

## test-files/frontend4/src/components/Navbar.tsx

### Functions
- closeMenu()
- toggleMenu()

### Types
- EXPORTS: const Navbar: Component = () => {
- IMPORTS: solid-js

---

## test-files/frontend4/src/components/app/AppSummary.tsx

### Functions
- handleActionComplete(result?: { transcript?, summary?: string })
- handleActionError(error: string)

### Types
- EXPORTS: AppSummary;
- IMPORTS: solid-js, solid-js, solid-js/web, @modular-forms/solid, @modular-forms/solid, @kobalte/core/toast, ~/components/app/ResultsDisplay, ~/components/app/PerplexityStyleForm, ~icons/carbon/close, ~/lib/api

### Variables
- DEFAULT_PROVIDER_MODEL

---

## test-files/frontend4/src/components/app/PerplexityStyleForm.tsx

### Functions
- copySummary()
- copyTranscript()
- copyTranscriptWithMetadata()
- downloadTranscriptJson()

### Types
- EXPORTS: interface PerplexityStyleFormProps {, const PerplexityStyleForm: Component<PerplexityStyleFormProps> = (props) => {, PerplexityStyleForm;
- IMPORTS: solid-js, solid-js, @kobalte/core/text-field, @kobalte/core/toggle-button, @kobalte/core/select, @kobalte/core/button, @kobalte/core/dropdown-menu, @kobalte/core/tooltip, @kobalte/core/toast, @solid-primitives/media, ~icons/carbon/ai-recommend, ~icons/carbon/copy-file, ~icons/carbon/checkmark, ~icons/carbon/close, ~icons/carbon/send, ~icons/carbon/text-wrap, @modular-forms/solid, ~/lib/schema, ~/lib/api

### Variables
- findOption

---

## test-files/frontend4/src/components/app/ResultsDisplay.tsx

### Functions
- resetInternalState(preserveTranscript: boolean)

### Types
- EXPORTS: interface ResultsDisplayProps {, const ResultsDisplay: Component<ResultsDisplayProps> = (props) => {, ResultsDisplay;
- IMPORTS: solid-js, solid-js, marked, isomorphic-dompurify, @kobalte/core/toast, @solid-primitives/resize-observer, @solid-primitives/media, ~icons/carbon/close, ~/lib/schema, ~/components/app/TranscriptDisplay, ~/components/app/SummaryDisplay

---

## test-files/frontend4/src/components/app/SummaryDisplay.tsx

### Types
- EXPORTS: const SummaryDisplay: Component<SummaryDisplayProps> = (props) => {
- IMPORTS: solid-js, solid-js, ~/lib/api

---

## test-files/frontend4/src/components/app/TranscriptDisplay.tsx

### Types
- EXPORTS: const TranscriptDisplay: Component<TranscriptDisplayProps> = (props) => {
- IMPORTS: solid-js, solid-js, ~/lib/api

---

## test-files/frontend4/src/lib/api.ts

### Functions
- extractVideoId(input: string)
- getApiBaseUrl()

### Types
- EXPORTS: function getApiBaseUrl(): string {, function extractVideoId(input: string): string | null {, interface TranscriptEntry {, interface EnhancedTranscriptResponse {, interface SummaryResponse {, interface TranscriptSummarizeResponse {, async function fetchTranscript(, async function fetchTranscriptAndSummary(, function streamTranscriptAndSummary(, interface FetchEventSourceController {, function streamSummaryOnly(
- IMPORTS: ./schema

---

## test-files/frontend4/src/lib/schema.ts

### Functions
- getLanguageDisplayName(code: Language)
- getModelDisplayName(code: LLMModel)
- getProviderDisplayName(code: LLMProvider)
- parseProviderModel(value: string | undefined | null)
- validateYoutubeUrl(url: string)

### Types
- EXPORTS: const youtubeUrlStringSchema = v.pipe(, const youtubeUrlSchema = v.object({ // Keep original object schema if needed elsewhere, function validateYoutubeUrl(url: string) {, const languageSchema = v.picklist(['en', 'es', 'fr', 'de', 'it', 'ja', 'ko', 'pt', 'ru', 'zh'], 'Please select a language');, type Language = v.InferOutput<typeof languageSchema>;, const languageDisplayNames: Record<Language, string> = {, function getLanguageDisplayName(code: Language): string {, const llmProviderSchema = v.picklist(['google', 'openai', 'groq', 'openrouter', 'auto'], 'Please select a provider'); // Add openai, groq, auto, type LLMProvider = v.InferOutput<typeof llmProviderSchema>;, const providerDisplayNames: Record<LLMProvider, string> = {, function getProviderDisplayName(code: LLMProvider): string {, const llmModelSchema = v.optional(v.picklist([, type LLMModel = v.InferOutput<typeof llmModelSchema>;, const modelDisplayNames: Record<Exclude<LLMModel, undefined>, string> = {, function getModelDisplayName(code: LLMModel): string {, interface ProviderModelOption {, const allProviderModelOptions: ProviderModelOption[] = [, const providerModelSchema = v.optional(, type ProviderModelValue = v.InferOutput<typeof providerModelSchema>; // Type will be string | undefined, function parseProviderModel(value: string | undefined | null): { provider?: LLMProvider; model?: LLMModel } | null {, const TranscriptionFormSchema = v.object({, type TranscriptionFormType = v.InferInput<typeof TranscriptionFormSchema>;
- IMPORTS: valibot, ./api

---

