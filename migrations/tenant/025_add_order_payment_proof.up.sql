-- Manual-payment proof: the customer uploads a transfer screenshot on the pending order page; the
-- admin reviews it next to the confirm button. Stores the object-storage key of the uploaded image.
ALTER TABLE orders
    ADD COLUMN payment_proof_s3_key VARCHAR(512);
