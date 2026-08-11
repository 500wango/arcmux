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
import { Check, Copy } from 'lucide-react'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'

import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'

type LanguageKey = 'curl' | 'python' | 'typescript' | 'go'

export function CodeIntegrationTabs({ className }: { className?: string }) {
  const { t } = useTranslation()
  const [activeTab, setActiveTab] = useState<LanguageKey>('curl')
  const [copied, setCopied] = useState(false)

  const codeSnippets: Record<LanguageKey, { label: string; code: string }> = {
    curl: {
      label: 'cURL',
      code: `curl -X POST https://api.arcmux.com/v1/chat/completions \\
  -H "Authorization: Bearer sk-your-api-key" \\
  -H "Content-Type: application/json" \\
  -d '{
    "model": "gpt-4o",
    "messages": [
      {"role": "system", "content": "You are a helpful assistant."},
      {"role": "user", "content": "Explain quantum computing in 2 sentences."}
    ],
    "stream": true
  }'`,
    },
    python: {
      label: 'Python',
      code: `from openai import OpenAI

client = OpenAI(
    api_key="sk-your-api-key",
    base_url="https://api.arcmux.com/v1"
)

response = client.chat.completions.create(
    model="claude-3-7-sonnet",
    messages=[
        {"role": "user", "content": "Write a high-performance LRU cache in Go."}
    ],
    stream=True
)

for chunk in response:
    if chunk.choices[0].delta.content:
        print(chunk.choices[0].delta.content, end="")`,
    },
    typescript: {
      label: 'TypeScript / Node',
      code: `import OpenAI from 'openai'

const client = new OpenAI({
  apiKey: process.env.ARCMUX_API_KEY,
  baseURL: 'https://api.arcmux.com/v1',
})

async function main() {
  const stream = await client.chat.completions.create({
    model: 'deepseek-r1',
    messages: [{ role: 'user', content: 'Design a distributed rate limiter.' }],
    stream: true,
  })

  for await (const chunk of stream) {
    process.stdout.write(chunk.choices[0]?.delta?.content || '')
  }
}

main()`,
    },
    go: {
      label: 'Go',
      code: `package main

import (
	"context"
	"fmt"
	"github.com/sashabaranov/go-openai"
)

func main() {
	cfg := openai.DefaultConfig("sk-your-api-key")
	cfg.BaseURL = "https://api.arcmux.com/v1"
	client := openai.NewClientWithConfig(cfg)

	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: "gpt-4o",
			Messages: []openai.ChatCompletionMessage{
				{Role: openai.ChatMessageRoleUser, Content: "Hello ArcMux!"},
			},
		},
	)
	if err != nil {
		panic(err)
	}
	fmt.Println(resp.Choices[0].Message.Content)
}`,
    },
  }

  const handleCopy = () => {
    navigator.clipboard.writeText(codeSnippets[activeTab].code)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  return (
    <div
      className={cn(
        'overflow-hidden rounded-2xl border border-border/80 bg-zinc-950 text-zinc-100 shadow-2xl backdrop-blur-xl',
        className
      )}
    >
      {/* Editor top bar */}
      <div className='flex flex-wrap items-center justify-between border-b border-zinc-800/80 bg-zinc-900/60 px-4 py-2.5'>
        <div className='flex items-center gap-2'>
          <div className='flex items-center gap-1.5'>
            <span className='size-3 rounded-full bg-red-500/80' />
            <span className='size-3 rounded-full bg-yellow-500/80' />
            <span className='size-3 rounded-full bg-green-500/80' />
          </div>
          <span className='ml-2 text-xs font-mono text-zinc-400'>
            {t('Universal Drop-in Integration')}
          </span>
        </div>

        <div className='flex items-center gap-1'>
          {(Object.keys(codeSnippets) as LanguageKey[]).map((lang) => (
            <button
              key={lang}
              type='button'
              onClick={() => setActiveTab(lang)}
              className={cn(
                'rounded-md px-2.5 py-1 text-xs font-mono font-medium transition-colors',
                activeTab === lang
                  ? 'bg-zinc-800 text-zinc-100 shadow-xs'
                  : 'text-zinc-400 hover:text-zinc-200'
              )}
            >
              {codeSnippets[lang].label}
            </button>
          ))}
          <Button
            size='icon-sm'
            variant='ghost'
            onClick={handleCopy}
            className='ml-2 text-zinc-400 hover:bg-zinc-800 hover:text-zinc-100'
            title={t('Copy code')}
          >
            {copied ? (
              <Check className='size-3.5 text-emerald-400' />
            ) : (
              <Copy className='size-3.5' />
            )}
          </Button>
        </div>
      </div>

      {/* Code view area */}
      <div className='relative overflow-x-auto p-5 font-mono text-xs leading-relaxed text-zinc-200'>
        <pre className='selection:bg-primary/30 selection:text-white'>
          <code>{codeSnippets[activeTab].code}</code>
        </pre>
      </div>
    </div>
  )
}
