import { uid } from "./model.js";

export function talkFromSession(session) {
  const messages = [];
  for (const turn of session?.turns || []) {
    messages.push({ id: uid(), role: "user", text: turn.user });
    if (turn.reply) messages.push({ id: uid(), role: "assistant", text: turn.reply, status: turn.status });
  }
  return { id: session.id, title: session.title || "会话", named: !!session.named, pending: false, loaded: true, messages };
}

export function cardTalk(card) {
  return { id: card.id, title: card.title || "会话", named: !!card.named, pending: false, loaded: false, messages: [] };
}
