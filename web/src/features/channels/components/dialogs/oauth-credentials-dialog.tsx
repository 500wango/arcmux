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
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  ClipboardCopy,
  ExternalLink,
  KeyRound,
  Loader2,
  MonitorSmartphone,
  Power,
  RefreshCw,
  Trash2,
  Upload,
} from 'lucide-react'
import { useEffect, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'

import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Checkbox } from '@/components/ui/checkbox'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Label } from '@/components/ui/label'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Textarea } from '@/components/ui/textarea'
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from '@/components/ui/tooltip'

import {
  completeUpstreamOAuth,
  deleteUpstreamOAuthCredential,
  importUpstreamOAuthCredentials,
  listUpstreamOAuthCredentials,
  pollUpstreamOAuth,
  refreshUpstreamOAuthCredential,
  setUpstreamOAuthCredentialEnabled,
  startUpstreamOAuth,
  type UpstreamOAuthCredential,
  type UpstreamOAuthFlowType,
  type UpstreamOAuthProvider,
  type UpstreamOAuthStartResult,
} from '../../api'
import {
  CHANNEL_TYPE_ANTHROPIC,
  CHANNEL_TYPE_CODEX,
  CHANNEL_TYPE_GEMINI,
  CHANNEL_TYPE_MOONSHOT,
  CHANNEL_TYPE_XAI,
} from '../../constants'

type OAuthCredentialsDialogProps = {
  open: boolean
  onOpenChange: (open: boolean) => void
  channelId: number
  channelType: number
}

type CredentialAction =
  | { type: 'refresh'; credential: UpstreamOAuthCredential }
  | { type: 'toggle'; credential: UpstreamOAuthCredential }
  | { type: 'delete'; credential: UpstreamOAuthCredential }

class LocalizedOAuthError extends Error {}

function oauthErrorMessage(error: unknown, fallback: string): string {
  return error instanceof LocalizedOAuthError ? error.message : fallback
}

function providersForChannelType(channelType: number): UpstreamOAuthProvider[] {
  if (channelType === CHANNEL_TYPE_CODEX) return ['codex']
  if (channelType === CHANNEL_TYPE_ANTHROPIC) return ['claude']
  if (channelType === CHANNEL_TYPE_GEMINI) return ['gemini-cli', 'antigravity']
  if (channelType === CHANNEL_TYPE_MOONSHOT) return ['kimi']
  if (channelType === CHANNEL_TYPE_XAI) return ['xai']
  return []
}

function formatTimestamp(timestamp?: number): string {
  if (!timestamp) return '-'
  return new Date(timestamp * 1000).toLocaleString()
}

export function OAuthCredentialsDialog(props: OAuthCredentialsDialogProps) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const providers = providersForChannelType(props.channelType)
  const [provider, setProvider] = useState<UpstreamOAuthProvider | null>(
    providers[0] ?? null
  )
  const [policyAccepted, setPolicyAccepted] = useState(false)
  const [activeSession, setActiveSession] =
    useState<UpstreamOAuthStartResult | null>(null)
  const [callbackUrl, setCallbackUrl] = useState('')
  const importInputRef = useRef<HTMLInputElement>(null)
  const [selectedCredentialIds, setSelectedCredentialIds] = useState<
    Set<number>
  >(() => new Set())

  useEffect(() => {
    setProvider(providersForChannelType(props.channelType)[0] ?? null)
    setActiveSession(null)
    setSelectedCredentialIds(new Set())
  }, [props.channelType])

  const credentialsQuery = useQuery({
    queryKey: ['channels', props.channelId, 'upstream-oauth-credentials'],
    queryFn: () => listUpstreamOAuthCredentials(props.channelId),
    enabled: props.open && props.channelId > 0,
  })

  const startMutation = useMutation({
    mutationFn: async (flowType: UpstreamOAuthFlowType) => {
      if (!provider) {
        throw new LocalizedOAuthError(
          t('OAuth is not supported for this channel')
        )
      }
      const response = await startUpstreamOAuth(
        props.channelId,
        provider,
        flowType,
        policyAccepted
      )
      if (!response.success || !response.data) {
        throw new LocalizedOAuthError(t('Failed to start OAuth authorization'))
      }
      return response.data
    },
    onSuccess: (session) => {
      setActiveSession(session)
      setCallbackUrl('')
      const target = session.authorization_url || session.verification_url
      if (target) window.open(target, '_blank', 'noopener,noreferrer')
    },
    onError: (error: unknown) =>
      toast.error(
        oauthErrorMessage(error, t('Failed to start OAuth authorization'))
      ),
  })

  const completeMutation = useMutation({
    mutationFn: async () => {
      if (!activeSession) {
        throw new LocalizedOAuthError(t('OAuth session is unavailable'))
      }
      const response = await completeUpstreamOAuth(
        props.channelId,
        activeSession.session_id,
        callbackUrl
      )
      if (!response.success) {
        throw new LocalizedOAuthError(
          t('Failed to complete OAuth authorization')
        )
      }
    },
    onSuccess: async () => {
      toast.success(t('OAuth account connected'))
      setActiveSession(null)
      setCallbackUrl('')
      await queryClient.invalidateQueries({
        queryKey: ['channels', props.channelId, 'upstream-oauth-credentials'],
      })
    },
    onError: (error: unknown) =>
      toast.error(
        oauthErrorMessage(error, t('Failed to complete OAuth authorization'))
      ),
  })

  const devicePollQuery = useQuery({
    queryKey: [
      'channels',
      props.channelId,
      'upstream-oauth-session',
      activeSession?.session_id,
    ],
    queryFn: async () => {
      if (!activeSession) {
        throw new LocalizedOAuthError(t('OAuth session is unavailable'))
      }
      const response = await pollUpstreamOAuth(
        props.channelId,
        activeSession.session_id
      )
      if (!response.success) {
        throw new LocalizedOAuthError(t('OAuth device authorization failed'))
      }
      return response.data
    },
    enabled:
      props.open &&
      activeSession?.flow_type === 'device' &&
      Boolean(activeSession.session_id),
    refetchInterval: (query) => {
      if (query.state.data?.status === 'completed') return false
      return (activeSession?.poll_interval ?? 5) * 1000
    },
    retry: false,
  })

  useEffect(() => {
    if (devicePollQuery.data?.status !== 'completed') return
    toast.success(t('OAuth account connected'))
    setActiveSession(null)
    void queryClient.invalidateQueries({
      queryKey: ['channels', props.channelId, 'upstream-oauth-credentials'],
    })
  }, [devicePollQuery.data?.status, props.channelId, queryClient, t])

  useEffect(() => {
    if (!devicePollQuery.error) return
    toast.error(
      oauthErrorMessage(
        devicePollQuery.error,
        t('OAuth device authorization failed')
      )
    )
  }, [devicePollQuery.error, t])

  const credentialAction = useMutation({
    mutationFn: async (action: CredentialAction) => {
      let response: { success: boolean; message?: string }
      if (action.type === 'refresh') {
        response = await refreshUpstreamOAuthCredential(
          props.channelId,
          action.credential.id
        )
      } else if (action.type === 'toggle') {
        response = await setUpstreamOAuthCredentialEnabled(
          props.channelId,
          action.credential.id,
          action.credential.status !== 1
        )
      } else {
        response = await deleteUpstreamOAuthCredential(
          props.channelId,
          action.credential.id
        )
      }
      if (!response.success) {
        throw new LocalizedOAuthError(t('OAuth credential operation failed'))
      }
      return action.type
    },
    onSuccess: async (actionType) => {
      toast.success(
        actionType === 'delete'
          ? t('OAuth account removed')
          : t('OAuth credential updated')
      )
      await queryClient.invalidateQueries({
        queryKey: ['channels', props.channelId, 'upstream-oauth-credentials'],
      })
    },
    onError: (error: unknown) =>
      toast.error(
        oauthErrorMessage(error, t('OAuth credential operation failed'))
      ),
  })

  const importMutation = useMutation({
    mutationFn: async (files: File[]) => {
      if (!provider) {
        throw new LocalizedOAuthError(
          t('OAuth is not supported for this channel')
        )
      }
      const totalSize = files.reduce((sum, file) => sum + file.size, 0)
      const importMaxBytes = credentialsQuery.data?.data?.import_max_bytes ?? 0
      if (importMaxBytes > 0 && totalSize > importMaxBytes) {
        throw new LocalizedOAuthError(
          t('OAuth JSON files must be {{size}} or smaller in total', {
            size: `${importMaxBytes / 1024 / 1024} MiB`,
          })
        )
      }
      const contents = await Promise.all(files.map((file) => file.text()))
      const accounts = contents.map((content, index) => {
        try {
          return JSON.parse(content) as unknown
        } catch {
          throw new LocalizedOAuthError(
            t('Invalid OAuth JSON file: {{name}}', {
              name: files[index]?.name ?? '',
            })
          )
        }
      })
      const response = await importUpstreamOAuthCredentials(
        props.channelId,
        provider,
        JSON.stringify({ accounts }),
        policyAccepted
      )
      if (!response.success || !response.data) {
        throw new LocalizedOAuthError(t('OAuth JSON import failed'))
      }
      return response.data.imported
    },
    onSuccess: async (count) => {
      toast.success(t('Imported {{count}} OAuth accounts', { count }))
      await queryClient.invalidateQueries({
        queryKey: ['channels', props.channelId, 'upstream-oauth-credentials'],
      })
    },
    onError: (error: unknown) =>
      toast.error(oauthErrorMessage(error, t('OAuth JSON import failed'))),
  })

  const batchDeleteMutation = useMutation({
    mutationFn: async (credentialIds: number[]) => {
      const responses = await Promise.all(
        credentialIds.map((credentialId) =>
          deleteUpstreamOAuthCredential(props.channelId, credentialId)
        )
      )
      if (responses.some((response) => !response.success)) {
        throw new LocalizedOAuthError(t('OAuth credential operation failed'))
      }
    },
    onSuccess: async () => {
      toast.success(t('OAuth account removed'))
      setSelectedCredentialIds(new Set())
      await queryClient.invalidateQueries({
        queryKey: ['channels', props.channelId, 'upstream-oauth-credentials'],
      })
    },
    onError: (error: unknown) =>
      toast.error(
        oauthErrorMessage(error, t('OAuth credential operation failed'))
      ),
  })

  const credentials = credentialsQuery.data?.data?.items ?? []
  const encryptionConfigured =
    credentialsQuery.data?.data?.encryption_configured ?? false
  const providerEnabled = provider
    ? (credentialsQuery.data?.data?.enabled_providers ?? []).includes(provider)
    : false
  const canAuthorize =
    Boolean(provider) &&
    policyAccepted &&
    encryptionConfigured &&
    providerEnabled &&
    !startMutation.isPending
  const canImport =
    Boolean(provider) &&
    policyAccepted &&
    encryptionConfigured &&
    providerEnabled &&
    !importMutation.isPending
  const copyDeviceCode = async () => {
    if (!activeSession?.user_code) return
    await navigator.clipboard.writeText(activeSession.user_code)
    toast.success(t('Device code copied'))
  }

  return (
    <Dialog open={props.open} onOpenChange={props.onOpenChange}>
      <DialogContent className='max-h-[calc(100dvh-2rem)] sm:max-w-2xl'>
        <DialogHeader>
          <DialogTitle>{t('Upstream OAuth accounts')}</DialogTitle>
          <DialogDescription>
            {t('Manage multiple OAuth accounts for this upstream provider.')}
          </DialogDescription>
        </DialogHeader>

        <ScrollArea className='max-h-[calc(100dvh-15rem)] pr-3'>
          <div className='space-y-4'>
            {!encryptionConfigured && !credentialsQuery.isLoading && (
              <Alert variant='destructive'>
                <AlertTitle>
                  {t('Credential encryption is not configured')}
                </AlertTitle>
                <AlertDescription>
                  {t(
                    'Set UPSTREAM_CREDENTIAL_ENCRYPTION_KEY before connecting OAuth accounts.'
                  )}
                </AlertDescription>
              </Alert>
            )}

            {encryptionConfigured &&
              !providerEnabled &&
              !credentialsQuery.isLoading && (
                <Alert>
                  <AlertTitle>
                    {t('Provider disabled by deployment policy')}
                  </AlertTitle>
                  <AlertDescription>
                    {t(
                      'Add this provider to UPSTREAM_OAUTH_ENABLED_PROVIDERS before authorizing it.'
                    )}
                  </AlertDescription>
                </Alert>
              )}

            <div className='flex items-start gap-3 rounded-md border p-3'>
              <Checkbox
                id='oauth-commercial-policy'
                checked={policyAccepted}
                onCheckedChange={(checked) =>
                  setPolicyAccepted(checked === true)
                }
              />
              <Label htmlFor='oauth-commercial-policy' className='leading-5'>
                {t(
                  'I have verified the provider terms and accept responsibility for commercial relay use.'
                )}
              </Label>
            </div>

            {providers.length > 1 && (
              <div className='space-y-2'>
                <Label>{t('OAuth provider')}</Label>
                <div className='flex flex-wrap gap-2'>
                  {providers.map((option) => (
                    <Button
                      key={option}
                      type='button'
                      variant={provider === option ? 'default' : 'outline'}
                      size='sm'
                      onClick={() => setProvider(option)}
                    >
                      {option === 'gemini-cli'
                        ? t('Gemini CLI')
                        : t('Antigravity')}
                    </Button>
                  ))}
                </div>
              </div>
            )}

            <div className='space-y-3'>
              <input
                ref={importInputRef}
                type='file'
                accept='.json,application/json'
                multiple
                className='hidden'
                onChange={(event) => {
                  const files = [...(event.target.files ?? [])]
                  event.target.value = ''
                  if (files.length > 0) importMutation.mutate(files)
                }}
              />
              <Button
                type='button'
                disabled={!canImport}
                onClick={() => importInputRef.current?.click()}
              >
                {importMutation.isPending ? (
                  <Loader2
                    className='h-4 w-4 animate-spin'
                    aria-hidden='true'
                  />
                ) : (
                  <Upload className='h-4 w-4' aria-hidden='true' />
                )}
                {t('Import OAuth JSON')}
              </Button>
              <div className='flex flex-wrap gap-2 border-t pt-3'>
                <Button
                  type='button'
                  variant='outline'
                  disabled={!canAuthorize}
                  onClick={() => startMutation.mutate('browser')}
                >
                  {startMutation.isPending ? (
                    <Loader2
                      className='h-4 w-4 animate-spin'
                      aria-hidden='true'
                    />
                  ) : (
                    <ExternalLink className='h-4 w-4' aria-hidden='true' />
                  )}
                  {t('Browser authorization')}
                </Button>
                {(provider === 'codex' ||
                  provider === 'kimi' ||
                  provider === 'xai') && (
                  <Button
                    type='button'
                    variant='outline'
                    disabled={!canAuthorize}
                    onClick={() => startMutation.mutate('device')}
                  >
                    <MonitorSmartphone className='h-4 w-4' aria-hidden='true' />
                    {t('Device authorization')}
                  </Button>
                )}
              </div>
            </div>

            {activeSession?.flow_type === 'browser' && (
              <div className='space-y-2 border-t pt-4'>
                <Label htmlFor='oauth-callback-url'>
                  {t('OAuth callback URL')}
                </Label>
                <Textarea
                  id='oauth-callback-url'
                  value={callbackUrl}
                  onChange={(event) => setCallbackUrl(event.target.value)}
                  placeholder={t(
                    'Paste the complete localhost callback URL here'
                  )}
                  rows={3}
                />
                <Button
                  type='button'
                  onClick={() => completeMutation.mutate()}
                  disabled={!callbackUrl.trim() || completeMutation.isPending}
                >
                  {completeMutation.isPending && (
                    <Loader2
                      className='h-4 w-4 animate-spin'
                      aria-hidden='true'
                    />
                  )}
                  {t('Complete authorization')}
                </Button>
              </div>
            )}

            {activeSession?.flow_type === 'device' && (
              <div className='space-y-3 border-t pt-4'>
                <p className='text-muted-foreground text-sm'>
                  {t(
                    'Enter this code on the provider device page. Authorization status updates automatically.'
                  )}
                </p>
                <div className='flex items-center gap-2'>
                  <code className='bg-muted rounded-md px-3 py-2 text-base font-semibold'>
                    {activeSession.user_code}
                  </code>
                  <Tooltip>
                    <TooltipTrigger
                      render={
                        <Button
                          type='button'
                          variant='outline'
                          size='icon'
                          onClick={copyDeviceCode}
                        />
                      }
                    >
                      <ClipboardCopy className='h-4 w-4' aria-hidden='true' />
                    </TooltipTrigger>
                    <TooltipContent>{t('Copy device code')}</TooltipContent>
                  </Tooltip>
                  {devicePollQuery.isFetching && (
                    <Loader2
                      className='text-muted-foreground h-4 w-4 animate-spin'
                      aria-label={t('Checking authorization status')}
                    />
                  )}
                </div>
              </div>
            )}

            <div className='border-t pt-4'>
              <div className='mb-2 flex items-center justify-between'>
                <div className='flex items-center gap-2'>
                  <Checkbox
                    checked={
                      credentials.length > 0 &&
                      selectedCredentialIds.size === credentials.length
                    }
                    onCheckedChange={(checked) =>
                      setSelectedCredentialIds(
                        checked
                          ? new Set(
                              credentials.map((credential) => credential.id)
                            )
                          : new Set()
                      )
                    }
                    aria-label={t('Select all')}
                  />
                  <h3 className='text-sm font-medium'>
                    {t('Connected accounts')}
                  </h3>
                  <Badge variant='outline'>{credentials.length}</Badge>
                </div>
                {selectedCredentialIds.size > 0 && (
                  <Button
                    type='button'
                    variant='destructive'
                    size='sm'
                    disabled={batchDeleteMutation.isPending}
                    onClick={() => {
                      if (!window.confirm(t('Confirm delete'))) return
                      batchDeleteMutation.mutate([...selectedCredentialIds])
                    }}
                  >
                    {batchDeleteMutation.isPending ? (
                      <Loader2
                        className='h-4 w-4 animate-spin'
                        aria-hidden='true'
                      />
                    ) : (
                      <Trash2 className='h-4 w-4' aria-hidden='true' />
                    )}
                    {t('Remove account')} ({selectedCredentialIds.size})
                  </Button>
                )}
              </div>
              {credentialsQuery.isLoading && (
                <div className='text-muted-foreground flex h-20 items-center justify-center'>
                  <Loader2
                    className='h-4 w-4 animate-spin'
                    aria-hidden='true'
                  />
                </div>
              )}
              {!credentialsQuery.isLoading && credentials.length === 0 && (
                <div className='text-muted-foreground flex h-20 items-center justify-center gap-2 rounded-md border border-dashed text-sm'>
                  <KeyRound className='h-4 w-4' aria-hidden='true' />
                  {t('No OAuth accounts connected')}
                </div>
              )}
              {!credentialsQuery.isLoading && credentials.length > 0 && (
                <div className='divide-y rounded-md border'>
                  {credentials.map((credential) => (
                    <div
                      key={credential.id}
                      className='flex flex-col gap-3 p-3 sm:flex-row sm:items-center sm:justify-between'
                    >
                      <div className='flex min-w-0 items-start gap-3'>
                        <Checkbox
                          checked={selectedCredentialIds.has(credential.id)}
                          onCheckedChange={(checked) =>
                            setSelectedCredentialIds((current) => {
                              const next = new Set(current)
                              if (checked) next.add(credential.id)
                              else next.delete(credential.id)
                              return next
                            })
                          }
                          aria-label={
                            credential.account_email || credential.account_id
                          }
                        />
                        <div className='min-w-0'>
                          <div className='flex flex-wrap items-center gap-2'>
                            <span className='truncate font-medium'>
                              {credential.account_email ||
                                credential.account_id}
                            </span>
                            <Badge
                              variant={
                                credential.status === 1
                                  ? 'default'
                                  : 'secondary'
                              }
                            >
                              {credential.status === 1
                                ? t('Enabled')
                                : t('Disabled')}
                            </Badge>
                          </div>
                          <p className='text-muted-foreground mt-1 text-xs'>
                            {t('Expires')}:{' '}
                            {formatTimestamp(credential.expires_at)}
                            {' · '}
                            {t('Failures')}: {credential.failure_count}
                          </p>
                        </div>
                      </div>
                      <div className='flex shrink-0 gap-1'>
                        <Tooltip>
                          <TooltipTrigger
                            render={
                              <Button
                                type='button'
                                variant='ghost'
                                size='icon-sm'
                                disabled={credentialAction.isPending}
                                onClick={() =>
                                  credentialAction.mutate({
                                    type: 'refresh',
                                    credential,
                                  })
                                }
                              />
                            }
                          >
                            <RefreshCw className='h-4 w-4' aria-hidden='true' />
                          </TooltipTrigger>
                          <TooltipContent>
                            {t('Refresh credential')}
                          </TooltipContent>
                        </Tooltip>
                        <Tooltip>
                          <TooltipTrigger
                            render={
                              <Button
                                type='button'
                                variant='ghost'
                                size='icon-sm'
                                disabled={credentialAction.isPending}
                                onClick={() =>
                                  credentialAction.mutate({
                                    type: 'toggle',
                                    credential,
                                  })
                                }
                              />
                            }
                          >
                            <Power className='h-4 w-4' aria-hidden='true' />
                          </TooltipTrigger>
                          <TooltipContent>
                            {credential.status === 1
                              ? t('Disable account')
                              : t('Enable account')}
                          </TooltipContent>
                        </Tooltip>
                        <Tooltip>
                          <TooltipTrigger
                            render={
                              <Button
                                type='button'
                                variant='ghost'
                                size='icon-sm'
                                disabled={credentialAction.isPending}
                                onClick={() => {
                                  if (
                                    window.confirm(
                                      t('Remove this OAuth account?')
                                    )
                                  ) {
                                    credentialAction.mutate({
                                      type: 'delete',
                                      credential,
                                    })
                                  }
                                }}
                              />
                            }
                          >
                            <Trash2
                              className='text-destructive h-4 w-4'
                              aria-hidden='true'
                            />
                          </TooltipTrigger>
                          <TooltipContent>{t('Remove account')}</TooltipContent>
                        </Tooltip>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>
          </div>
        </ScrollArea>

        <DialogFooter>
          <Button
            type='button'
            variant='outline'
            onClick={() => props.onOpenChange(false)}
          >
            {t('Close')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
