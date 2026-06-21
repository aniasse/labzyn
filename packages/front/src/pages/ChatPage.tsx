import { useRef, useState, useEffect } from 'react';
import { useMutation, useQuery } from '@tanstack/react-query';
import { Send, Sparkles, Wrench } from 'lucide-react';
import { api, type AgentStep } from '@/lib/api';
import { Button } from '@/components/ui/button';

type ChatMessage = {
  role: 'user' | 'assistant';
  content: string;
  steps?: AgentStep[];
};

export function ChatPage() {
  const { data: status } = useQuery({ queryKey: ['agentStatus'], queryFn: api.agentStatus });
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [input, setInput] = useState('');
  const endRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    endRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages]);

  const send = useMutation({
    mutationFn: (history: ChatMessage[]) =>
      api.agentChat(history.map((m) => ({ role: m.role, content: m.content }))),
    onSuccess: (res) =>
      setMessages((m) => [...m, { role: 'assistant', content: res.reply, steps: res.steps }]),
    onError: (e: Error) =>
      setMessages((m) => [...m, { role: 'assistant', content: `⚠ ${e.message}` }]),
  });

  const submit = () => {
    const text = input.trim();
    if (!text || send.isPending) return;
    const next: ChatMessage[] = [...messages, { role: 'user', content: text }];
    setMessages(next);
    setInput('');
    send.mutate(next);
  };

  const disabled = status && !status.enabled;

  return (
    <div className="mx-auto flex h-full max-w-2xl flex-col">
      <div className="flex items-center gap-2 border-b border-border px-5 py-3">
        <Sparkles size={18} className="text-primary" />
        <h1 className="text-md font-semibold text-foreground">Ask Labzyn</h1>
      </div>

      <div className="min-h-0 flex-1 overflow-auto px-5 py-4">
        {messages.length === 0 && (
          <div className="mt-10 text-center">
            <p className="text-sm font-medium text-foreground">Your research copilot</p>
            <p className="mt-1 text-sm text-tertiary">
              Ask about your data, create records, or submit an HPC job.
            </p>
          </div>
        )}

        <div className="flex flex-col gap-4">
          {messages.map((m, i) => (
            <Bubble key={i} message={m} />
          ))}
          {send.isPending && <div className="text-sm text-tertiary">Labzyn is thinking…</div>}
          <div ref={endRef} />
        </div>
      </div>

      <div className="border-t border-border px-5 py-3">
        {disabled ? (
          <p className="rounded-sm bg-muted px-3 py-2 text-xs text-tertiary">
            Agent not configured. Set <code>LABZYN_LLM_API_KEY</code> (and base URL/model) on the
            server to enable chat.
          </p>
        ) : (
          <div className="flex items-end gap-2">
            <textarea
              value={input}
              onChange={(e) => setInput(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === 'Enter' && !e.shiftKey) {
                  e.preventDefault();
                  submit();
                }
              }}
              rows={1}
              placeholder="Message Labzyn…"
              className="max-h-32 flex-1 resize-none rounded-sm border border-border bg-background px-3 py-2 text-sm text-foreground outline-none focus:border-primary"
            />
            <Button variant="primary" className="h-9" onClick={submit} disabled={send.isPending}>
              <Send size={15} />
            </Button>
          </div>
        )}
      </div>
    </div>
  );
}

function Bubble({ message }: { message: ChatMessage }) {
  const isUser = message.role === 'user';
  return (
    <div className={isUser ? 'flex justify-end' : 'flex justify-start'}>
      <div
        className={
          isUser
            ? 'max-w-[80%] rounded-md bg-primary px-3 py-2 text-sm text-primary-foreground'
            : 'max-w-[80%] rounded-md bg-muted px-3 py-2 text-sm text-foreground'
        }
      >
        {message.steps && message.steps.length > 0 && (
          <div className="mb-2 flex flex-wrap gap-1">
            {message.steps.map((s, i) => (
              <span
                key={i}
                className="inline-flex items-center gap-1 rounded-pill bg-background px-2 py-0.5 text-xxs text-secondary"
                title={s.result}
              >
                <Wrench size={11} />
                {s.tool}
              </span>
            ))}
          </div>
        )}
        <span className="whitespace-pre-wrap">{message.content}</span>
      </div>
    </div>
  );
}
