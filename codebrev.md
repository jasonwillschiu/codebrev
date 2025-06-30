# Code Structure Outline

This file provides an overview of available functions, types, and variables per file for LLM context.

## test-files/frontend4/src/components/FAQSection.astro

---

## test-files/frontend4/src/components/FeaturesSection.astro

---

## test-files/frontend4/src/components/GoogleAnalytics.astro

### Types
- ASTRO_PROPS: id

### Variables
- gaId
- isProduction

---

## test-files/frontend4/src/components/Navbar.tsx

### Functions
- closeMenu()
- toggleMenu()

### Types
- EXPORTS: const Navbar: Component = () => {
- IMPORTS: solid-js
- JSX_COMPONENTS: For, Show
- component Navbar
- interface NavLinkItem

---

## test-files/frontend4/src/components/OptimizedImage.astro

### Types
- ASTRO_COMPONENTS: Picture
- ASTRO_IMPORTS: astro:assets

---

## test-files/frontend4/src/components/PricingSection.astro

---

## test-files/frontend4/src/components/TestimonialsSection.astro

---

## test-files/frontend4/src/components/app/AppSummary.tsx

### Functions
- handleActionComplete(result?: { transcript?, summary?: string })
- handleActionError(error: string)

### Types
- EXPORTS: AppSummary;
- IMPORTS: solid-js, solid-js, solid-js/web, @modular-forms/solid, @modular-forms/solid, @kobalte/core/toast, ~/components/app/ResultsDisplay, ~/components/app/PerplexityStyleForm, ~icons/carbon/close, ~/lib/api
- JSX_COMPONENTS: TranscriptionFormType, Omit, ResultsDisplayProps, EnhancedTranscriptResponse, CompletedParams, Toast, CarbonClose, T, Show, Portal, Suspense, LazyPerplexityStyleForm, LazyResultsDisplay
- component AppSummary
- interface CompletedParams
- interface SelectOption

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
- JSX_COMPONENTS: T, TranscriptionFormType, EnhancedTranscriptResponse, PerplexityStyleFormProps, Toast, CarbonClose, Form, Field, KobalteTextField, Show, KobalteTooltip, KobalteToggleButton, KobalteSelect, SelectOption, CarbonCheckmark, TablerPhotoAi, CarbonAiRecommend, CarbonTextWrap, KobalteDropdownMenu, CarbonCopyFile, KobalteButton, CarbonSend
- component PerplexityStyleForm
- interface PerplexityStyleFormProps
- interface SelectOption
- type SpecificField
- type SpecificForm

### Variables
- findOption

---

## test-files/frontend4/src/components/app/ResultsDisplay.tsx

### Functions
- resetInternalState(preserveTranscript: boolean)

### Types
- EXPORTS: interface ResultsDisplayProps {, const ResultsDisplay: Component<ResultsDisplayProps> = (props) => {, ResultsDisplay;
- IMPORTS: solid-js, solid-js, marked, isomorphic-dompurify, @kobalte/core/toast, @solid-primitives/resize-observer, @solid-primitives/media, ~icons/carbon/close, ~/lib/schema, ~/components/app/TranscriptDisplay, ~/components/app/SummaryDisplay
- JSX_COMPONENTS: ResultsDisplayProps, EnhancedTranscriptResponse, Partial, SummaryResponse, EventSource, HTMLDivElement, Toast, CarbonClose, Show, TranscriptDisplay, SummaryDisplay
- component ResultsDisplay
- interface ResultsDisplayProps

---

## test-files/frontend4/src/components/app/SummaryDisplay.tsx

### Types
- EXPORTS: const SummaryDisplay: Component<SummaryDisplayProps> = (props) => {
- IMPORTS: solid-js, solid-js, ~/lib/api
- JSX_COMPONENTS: SummaryResponse, SummaryDisplayProps, Show
- component SummaryDisplay
- interface SummaryDisplayProps

---

## test-files/frontend4/src/components/app/TranscriptDisplay.tsx

### Types
- EXPORTS: const TranscriptDisplay: Component<TranscriptDisplayProps> = (props) => {
- IMPORTS: solid-js, solid-js, ~/lib/api
- JSX_COMPONENTS: TranscriptDisplayProps, NormalizedEntry, Show
- component TranscriptDisplay
- interface ChunkedEntry
- interface NormalizedEntry
- interface TranscriptDisplayProps

---

## test-files/frontend4/src/layouts/Base.astro

### Types
- ASTRO_COMPONENTS: GoogleAnalytics, Navbar
- ASTRO_IMPORTS: ../components/Navbar.tsx, ../components/GoogleAnalytics.astro
- CLIENT_DIRECTIVES: idle
- SLOTS: default

---

## test-files/frontend4/src/layouts/Layout.astro

### Types
- SLOTS: default

---

## test-files/frontend4/src/layouts/MarkdownLayout.astro

### Types
- ASTRO_COMPONENTS: Base
- ASTRO_IMPORTS: ./Base.astro
- ASTRO_PROPS: frontmatter
- SLOTS: default

---

## test-files/frontend4/src/lib/api.ts

### Functions
- extractVideoId(input: string)
- getApiBaseUrl()

### Types
- EXPORTS: function getApiBaseUrl(): string {, function extractVideoId(input: string): string | null {, interface TranscriptEntry {, interface EnhancedTranscriptResponse {, interface SummaryResponse {, interface TranscriptSummarizeResponse {, async function fetchTranscript(, async function fetchTranscriptAndSummary(, function streamTranscriptAndSummary(, interface FetchEventSourceController {, function streamSummaryOnly(
- IMPORTS: ./schema
- JSX_COMPONENTS: EnhancedTranscriptResponse, TranscriptSummarizeResponse
- interface EnhancedTranscriptResponse
- interface FetchEventSourceController
- interface StreamSummaryRequestBody
- interface SummaryResponse
- interface TranscriptEntry
- interface TranscriptSummarizeResponse

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
- JSX_COMPONENTS: Language, LLMProvider, Exclude, LLMModel
- interface ProviderModelOption
- type LLMModel
- type LLMProvider
- type Language
- type ProviderModelValue
- type TranscriptionFormType

---

## test-files/frontend4/src/pages/app.astro

### Types
- ASTRO_COMPONENTS: Base, AppSummary
- ASTRO_IMPORTS: ~/layouts/Base.astro, ~/components/app/AppSummary.tsx
- CLIENT_DIRECTIVES: only

---

## test-files/frontend4/src/pages/blog/index.astro

### Types
- ASTRO_COMPONENTS: Base
- ASTRO_IMPORTS: ../../layouts/Base.astro

---

## test-files/frontend4/src/pages/features.astro

### Types
- ASTRO_COMPONENTS: Base, FeaturesSection, TestimonialsSection
- ASTRO_IMPORTS: ../layouts/Base.astro, ../components/FeaturesSection.astro, ../components/TestimonialsSection.astro

---

## test-files/frontend4/src/pages/index.astro

### Types
- ASTRO_COMPONENTS: Base
- ASTRO_IMPORTS: ../layouts/Base.astro

---

## test-files/frontend4/src/pages/pricing.astro

### Types
- ASTRO_COMPONENTS: Base, PricingSection, TestimonialsSection, FAQSection
- ASTRO_IMPORTS: ../layouts/Base.astro, ../components/PricingSection.astro, ../components/FAQSection.astro, ../components/TestimonialsSection.astro

---

