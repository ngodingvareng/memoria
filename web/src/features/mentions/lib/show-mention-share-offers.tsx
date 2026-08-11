import { toast } from 'sonner';
import type { MentionShareOffer } from './sync-mentions';
import { MentionShareOfferToast } from '../components/mention-share-offer-toast';

export function showMentionShareOffers(
  momentId: string,
  offers: MentionShareOffer[]
) {
  for (const offer of offers) {
    toast.custom((toastId) => (
      <MentionShareOfferToast
        momentId={momentId}
        offer={offer}
        toastId={toastId}
      />
    ));
  }
}
