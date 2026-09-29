"use client";

import Link from "next/link";
import { useState, useTransition } from "react";

import { Alert } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button, buttonClass } from "@/components/ui/button";
import { Dialog } from "@/components/ui/dialog";
import { textareaClass } from "@/components/ui/field";
import { Info, Pencil, Star, Trash } from "@/components/ui/icons";
import { AverageStars, StarRow } from "@/components/ui/stars";
import { toast } from "@/components/ui/toaster";
import { createReview, deleteReview, updateReview } from "@/features/reviews/actions";
import { formatReviewCount } from "@/lib/format";
import type { Review } from "@/types/api";

const WORDS = ["Select a rating", "Poor", "Fair", "Good", "Very good", "Excellent"];
const AVATAR_BG = ["#F2D8A7", "#CFE0D4", "#F6D2C4", "#DAD5EA", "#D8E4F0"];

const MONTHS = ["Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"];

/** "28 Sep 2026" (Intl's en-GB would print "Sept"). */
function reviewDate(iso: string) {
  const d = new Date(iso);
  return `${d.getDate()} ${MONTHS[d.getMonth()]} ${d.getFullYear()}`;
}

function Author({ review, mine }: { review: Review; mine?: boolean }) {
  const initial = review.author_name.trim().charAt(0).toUpperCase() || "?";
  const bg = mine ? "#245038" : AVATAR_BG[initial.charCodeAt(0) % AVATAR_BG.length];
  return (
    <div className="flex items-center gap-3">
      <span aria-hidden className="flex size-11 items-center justify-center rounded-full font-bold" style={{ background: bg, color: mine ? "#fff" : "#221C18" }}>
        {initial}
      </span>
      <div className="flex flex-col gap-0.5">
        <span className="flex items-center gap-2 text-base font-bold">
          {review.author_name}
          {mine && (
            <Badge tone="mine" className="h-6 text-xs">
              Your review
            </Badge>
          )}
        </span>
        <span className="text-sm text-ink-3">{reviewDate(review.created_at)}</span>
      </div>
    </div>
  );
}

type Mode = "guest" | "owner" | "customer";

type Props = {
  restaurantId: string;
  restaurantName: string;
  average: number | null;
  count: number;
  mine: Review | null;
  others: Review[];
  mode: Mode;
  loginHref: string;
  moreHref: string | null;
  openInitially?: boolean;
};

export function ReviewsSection({ restaurantId, restaurantName, average, count, mine, others, mode, loginHref, moreHref, openInitially }: Props) {
  const [formOpen, setFormOpen] = useState(Boolean(openInitially && mode === "customer"));
  const [rating, setRating] = useState(mine?.rating ?? 0);
  const [comment, setComment] = useState(mine?.comment ?? "");
  const [error, setError] = useState<string | null>(null);
  const [confirmDelete, setConfirmDelete] = useState(false);
  const [pending, startTransition] = useTransition();
  const editing = mine !== null;

  const openForm = () => {
    setRating(mine?.rating ?? 0);
    setComment(mine?.comment ?? "");
    setError(null);
    setFormOpen(true);
  };

  const submit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!rating) return;
    startTransition(async () => {
      const draft = { rating, comment };
      const result = editing ? await updateReview(mine!.id, draft) : await createReview(restaurantId, draft);
      if (result.status === "success") {
        setFormOpen(false);
        toast({ message: editing ? "Your review has been updated." : "Thanks! Your review has been posted." });
      } else if (result.status === "error") {
        setError(result.fieldErrors?.rating ?? result.fieldErrors?.comment ?? result.message);
      }
    });
  };

  const remove = () =>
    startTransition(async () => {
      const result = await deleteReview(mine!.id);
      setConfirmDelete(false);
      if (result.status === "success") {
        setRating(0);
        setComment("");
        toast({ message: "Your review has been deleted." });
      } else if (result.status === "error") {
        toast({ message: result.message, kind: "error" });
      }
    });

  let action: React.ReactNode = null;
  if (mode === "owner") {
    action = (
      <span className="flex h-10 items-center gap-2 rounded-xl bg-sand px-3.5 text-sm font-semibold text-body">
        <Info size={16} />
        Owners can&apos;t review their own restaurant
      </span>
    );
  } else if (mode === "guest") {
    action = (
      <a href={loginHref} className={buttonClass("secondary", "md", "h-11 px-[18px]")}>
        <Pencil size={18} />
        Write a Review
      </a>
    );
  } else if (!formOpen) {
    action = (
      <Button variant="secondary" className="h-11 px-[18px]" onClick={openForm}>
        <Pencil size={18} />
        {editing ? "Edit your review" : "Write a Review"}
      </Button>
    );
  }

  return (
    <section id="reviews" aria-labelledby="reviews-h" className="flex scroll-mt-24 flex-col gap-6 border-t border-line-mid pt-10">
      <div className="flex flex-wrap items-center justify-between gap-4">
        <h2 id="reviews-h" className="m-0 font-display text-[30px] font-bold tracking-[-0.02em]">
          Reviews
        </h2>
        {action}
      </div>

      {average !== null ? (
        <div className="flex items-center gap-6 rounded-[20px] border border-line-soft bg-surface px-7 py-6">
          <span className="font-display text-[64px] leading-none font-bold tracking-[-0.03em]">{average.toFixed(1)}</span>
          <div className="flex flex-col gap-2">
            <AverageStars average={average} />
            <span className="text-[15px] text-ink-2">
              Average of <strong className="text-ink">{formatReviewCount(count)}</strong>
            </span>
          </div>
        </div>
      ) : (
        !formOpen && (
          <div className="flex flex-col items-center gap-2.5 rounded-[20px] border-[1.5px] border-dashed border-line bg-surface px-6 py-12 text-center">
            <div className="flex gap-1">
              {[0, 1, 2, 3, 4].map((i) => (
                <Star key={i} size={28} fill="#E6DCCD" />
              ))}
            </div>
            <strong className="font-display text-[22px]">No reviews yet</strong>
            <span className="text-[15px] text-ink-2">
              {mode === "owner" ? "Reviews from diners will appear here." : "Been here? Be the first to share your experience."}
            </span>
          </div>
        )
      )}

      {formOpen && (
        <form onSubmit={submit} aria-labelledby="rv-h" className="flex flex-col gap-5 rounded-[20px] border-[1.5px] border-ink bg-surface p-5 sm:p-7">
          <h3 id="rv-h" className="m-0 text-xl font-bold">
            {editing ? "Edit your review" : "Write a Review"}
          </h3>
          <div className="flex flex-col gap-2.5">
            <span id="rv-rating-label" className="text-sm font-semibold">
              Your rating
            </span>
            <div className="flex flex-wrap items-center gap-4">
              <div role="radiogroup" aria-labelledby="rv-rating-label" className="flex gap-1">
                {[1, 2, 3, 4, 5].map((i) => (
                  <button
                    key={i}
                    type="button"
                    role="radio"
                    aria-checked={rating === i}
                    aria-label={`${i} ${i === 1 ? "star" : "stars"} — ${WORDS[i]}`}
                    onClick={() => setRating(i)}
                    className="flex size-11 cursor-pointer items-center justify-center rounded-[10px] hover:bg-hover"
                  >
                    <Star size={34} fill={i <= rating ? "#E3A327" : "#E6DCCD"} />
                  </button>
                ))}
              </div>
              <span aria-live="polite" className="text-[15px] font-semibold text-ink-2">
                {rating ? `${rating} / 5 · ${WORDS[rating]}` : WORDS[0]}
              </span>
            </div>
          </div>
          <div className="flex flex-col gap-1.5">
            <label htmlFor="rv-comment" className="text-sm font-semibold">
              Comment
            </label>
            <textarea
              id="rv-comment"
              rows={4}
              maxLength={2000}
              value={comment}
              onChange={(e) => setComment(e.target.value)}
              placeholder="What did you eat? How was the service and atmosphere?"
              className={textareaClass()}
            />
          </div>
          {error && <Alert tone="error">{error}</Alert>}
          <div className="flex justify-end gap-3">
            <Button variant="ghost" onClick={() => setFormOpen(false)} disabled={pending}>
              Cancel
            </Button>
            <Button type="submit" disabled={!rating || pending} className="px-6 font-semibold">
              {pending ? "Saving…" : editing ? "Save Review" : "Submit Review"}
            </Button>
          </div>
        </form>
      )}

      <div className="flex flex-col gap-4">
        {mine && !formOpen && (
          <article className="flex flex-col gap-3 rounded-[20px] border-[1.5px] border-peach-line bg-surface p-6">
            <div className="flex flex-wrap items-center justify-between gap-3">
              <Author review={mine} mine />
              <div className="flex gap-2">
                <Button variant="secondary" size="sm" onClick={openForm}>
                  <Pencil size={16} />
                  Edit
                </Button>
                <Button variant="dangerOutline" size="sm" onClick={() => setConfirmDelete(true)}>
                  <Trash size={16} />
                  Delete
                </Button>
              </div>
            </div>
            <StarRow rating={mine.rating} />
            {mine.comment && <p className="m-0 text-base leading-relaxed whitespace-pre-line text-body">{mine.comment}</p>}
          </article>
        )}
        {others.map((r) => (
          <article key={r.id} className="flex flex-col gap-3 rounded-[20px] border border-line-soft bg-surface p-6">
            <Author review={r} />
            <StarRow rating={r.rating} />
            {r.comment && <p className="m-0 text-base leading-relaxed whitespace-pre-line text-body">{r.comment}</p>}
          </article>
        ))}
        {moreHref && (
          <Link href={moreHref} scroll={false} className={buttonClass("secondary", "md", "self-center px-6")}>
            Show more reviews
          </Link>
        )}
      </div>

      <Dialog open={confirmDelete} onClose={() => setConfirmDelete(false)} role="alertdialog" labelledBy="del-h" describedBy="del-d">
        <h2 id="del-h" className="m-0 text-[22px] font-bold">
          Delete your review?
        </h2>
        <p id="del-d" className="m-0 text-base leading-[1.55] text-body">
          Your rating and comment will be removed from {restaurantName}. This can&apos;t be undone.
        </p>
        <div className="flex flex-col-reverse justify-end gap-3 pt-2 sm:flex-row">
          <Button variant="secondary" onClick={() => setConfirmDelete(false)} autoFocus>
            Keep review
          </Button>
          <Button variant="danger" onClick={remove} disabled={pending}>
            {pending ? "Deleting…" : "Delete review"}
          </Button>
        </div>
      </Dialog>
    </section>
  );
}
