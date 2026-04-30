-- SPDX-FileCopyrightText: 2026 Scott Friedman
-- SPDX-License-Identifier: Apache-2.0

-- Countries-of-concern awareness training module per NIH NOT-OD-25-083 (April 2025).
-- Completion writes attest:coc-check-current=true to the user's IAM role.
-- This module is recommended alongside nih-research-security for NIH-funded projects.

INSERT INTO training_modules (name, title, description, category, difficulty, estimated_minutes, content) VALUES

('countries-of-concern-awareness',
 'Countries-of-Concern Awareness (NOT-OD-25-083)',
 'Required for access to NIH controlled-access genomic data when institutional collaborators are involved. Covers the NIH April 2025 policy barring access from designated countries of concern, with a focus on institutional affiliation (not citizenship).',
 'compliance', 'intermediate', 30,
 '{"sections": [
  {
    "id": "policy-overview",
    "title": "The Policy: What NOT-OD-25-083 Requires",
    "type": "text",
    "content": "In April 2025, NIH issued NOT-OD-25-083 establishing that individuals whose primary institutional affiliation is in a designated country of concern may not access NIH controlled-access genomic data.\n\n**Designated countries of concern (as of 2025):** China (CN), Russia (RU), Iran (IR), North Korea (KP), Cuba (CU), Venezuela (VE).\n\n**Why this policy exists:** NIH determined that certain foreign government talent recruitment programs and data-sharing practices pose risks to the security and integrity of federally funded research. Controlled-access genomic data is particularly sensitive because of its potential for misuse.\n\n**Effective date:** April 2025. Institutions must comply when renewing or amending Data Use Agreements (DUAs). All new DUA applications are immediately subject to this policy.\n\n**What it means for your institution:** You must maintain records of the institutional affiliation country for all Approved Users and verify that no current or prospective Approved Users have a primary affiliation with an institution in a designated country."
  },
  {
    "id": "affiliation-vs-citizenship",
    "title": "Critical Distinction: Affiliation vs. Citizenship",
    "type": "text",
    "content": "The most important — and most commonly misunderstood — aspect of this policy is that it is based on **institutional affiliation**, not citizenship or nationality.\n\n**Examples:**\n\n- A US citizen whose primary appointment is at Peking University (China): **subject to the restriction**. Their institutional affiliation is Chinese, regardless of their citizenship.\n\n- A Chinese national employed full-time at a US university with no concurrent Chinese institutional appointment: **generally not subject to the restriction**. Their primary institutional affiliation is American.\n\n- A postdoctoral fellow at your institution who is a Chinese national but holds no position at a Chinese institution: **assess based on your institution's procedures** — the primary appointment is what matters.\n\n**Why this distinction matters:** NIH is concerned about institutional data-sharing agreements and talent recruitment programs, which are tied to institutional affiliation rather than individual nationality.\n\n**When in doubt:** Consult your institution''s Export Control office and Research Security officer. Do not make this determination unilaterally.\n\n**Concurrent appointments:** A researcher who holds appointments at both a US institution and a Chinese institution (e.g., a ''thousand talents'' honoree) is likely subject to the restriction. Disclosure of all concurrent appointments is required in the DUA process."
  },
  {
    "id": "institutional-obligations",
    "title": "What Your Institution Must Do",
    "type": "text",
    "content": "Your institution has specific obligations to comply with NOT-OD-25-083:\n\n**1. Maintain Approved User Records**\nKeep current records of the institutional affiliation country for every Approved User. This information must be updated when users change institutions and reviewed at each DUA renewal.\n\n**2. Conduct Countries-of-Concern Checks**\nBefore granting or renewing Approved User access to controlled-access data, verify that the user''s primary institutional affiliation is not in a designated country. Document this check. In the Provabl suite, this is recorded by setting attest:coc-check-current=true in the user''s IAM role tags after the check is performed.\n\n**3. Revoke Access When Necessary**\nIf an Approved User''s primary institutional affiliation changes to a designated country, access must be revoked promptly. The user may reapply once their primary affiliation is no longer in a designated country.\n\n**4. Report to NIH**\nIf you discover that a current Approved User has an undisclosed affiliation with a designated country, report this to NIH as a compliance incident.\n\n**5. Update Your DUA**\nNew DUAs and renewals must attest that all Approved Users have been checked against the countries-of-concern list. The IT contact signing the DUA bears responsibility for this attestation.\n\n**How attest enforces this at runtime:** The Cedar policy compiled from the nih-gds framework evaluates principal.institutional_affiliation_country and principal.countries_of_concern_check_current before every access to NIH controlled-access resources. If the check is not current, access is denied — fail closed."
  }
 ],
 "quiz": [
  {
    "id": "q1",
    "question": "A Chinese national is employed full-time as an Assistant Professor at your US university. They hold no position at any Chinese institution and have no active talent recruitment program participation. Under NOT-OD-25-083, are they subject to the countries-of-concern restriction?",
    "options": [
      "Yes — any Chinese national is barred from controlled-access data",
      "No — their primary institutional affiliation is the US university",
      "It depends only on whether they were born in China",
      "The policy does not apply to assistant professors"
    ],
    "correct": 1,
    "explanation": "NOT-OD-25-083 is based on institutional affiliation, not citizenship or nationality. This researcher''s primary institutional affiliation is the US university. However, if they have any concurrent appointment at a Chinese institution, that must be disclosed and evaluated separately."
  },
  {
    "id": "q2",
    "question": "A US citizen is an Approved User on your DUA. She then accepts an honorary professor position at a university in China. What must you do?",
    "options": [
      "Nothing — she is a US citizen, so the policy does not apply",
      "Notify her that she must disclose this on her next DUA renewal",
      "Evaluate whether this creates a concurrent institutional affiliation with a designated country; potentially revoke access pending institutional review",
      "Immediately revoke her access without further investigation"
    ],
    "correct": 2,
    "explanation": "An honorary professorship may or may not constitute a primary institutional affiliation — this requires evaluation by your Export Control and Research Security offices. The key question is whether her primary appointment is now in a designated country. Access should be placed on hold pending that determination, not automatically revoked without review."
  },
  {
    "id": "q3",
    "question": "Your institution uses the Provabl qualify system. After completing this training and the institutional countries-of-concern check, which IAM tag is written to your role?",
    "options": [
      "attest:itar-training=true",
      "attest:coc-check-current=true",
      "attest:country=US",
      "attest:nih-approval=true"
    ],
    "correct": 1,
    "explanation": "Completing this training and the associated institutional check writes attest:coc-check-current=true to your IAM role. The attest Cedar PDP requires this tag (in addition to active NIH approval) before granting access to NIH controlled-access genomic resources."
  },
  {
    "id": "q4",
    "question": "What is the primary purpose of the attest:country IAM tag written by your institution?",
    "options": [
      "To track which country your AWS account is hosted in",
      "To record your citizenship for ITAR export control purposes",
      "To record your institutional affiliation country for NIH GDS and ITAR policy enforcement",
      "To restrict your access to US-region AWS services only"
    ],
    "correct": 2,
    "explanation": "The attest:country tag records your primary institutional affiliation country (ISO 3166-1 alpha-2 code). The NIH GDS Cedar policy uses this to enforce NOT-OD-25-083, and the ITAR framework uses it for deemed-export access control. It reflects institutional affiliation, not citizenship."
  },
  {
    "id": "q5",
    "question": "If the countries-of-concern check has not been completed for a user (attest:coc-check-current=false), what happens when they attempt to access NIH controlled-access genomic data?",
    "options": [
      "Access is granted but a warning is logged",
      "Access is granted with a 24-hour grace period",
      "Access is denied — the Cedar policy fails closed when the check is not current",
      "Access is granted pending manual review by the compliance officer"
    ],
    "correct": 2,
    "explanation": "The NIH GDS Cedar policy fails closed: if attest:coc-check-current is false or absent, access is denied. This is intentional — uncertainty about a user''s affiliation status must be resolved before access is granted, not after. The compliance officer must complete the check and update the tag."
  }
 ],
 "passing_score": 80,
 "prerequisite_modules": ["nih-research-security"],
 "certificate_template": "coc-awareness-completion"
}')

ON CONFLICT (name) DO UPDATE SET
  title = EXCLUDED.title,
  description = EXCLUDED.description,
  category = EXCLUDED.category,
  difficulty = EXCLUDED.difficulty,
  estimated_minutes = EXCLUDED.estimated_minutes,
  content = EXCLUDED.content;
