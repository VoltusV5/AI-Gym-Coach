// Drafts survive navigation to sign-in in this tab and are never written to storage.
const drafts = new Map()

/** Keep a detached draft for the session that expired. */
export function stashDraft(owner, key, draft) {
  if (owner) drafts.set(`${owner}:${key}`, JSON.parse(JSON.stringify(draft)))
}

/** Consume a saved draft only after the original account signs back in. */
export function takeDraft(owner, key) {
  const id = `${owner}:${key}`
  const draft = drafts.get(id)
  drafts.delete(id)
  return draft
}

/** Remove all previous-account drafts after switching identities. */
export function clearDrafts() {
  drafts.clear()
}
