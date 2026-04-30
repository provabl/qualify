-- SPDX-FileCopyrightText: 2026 Scott Friedman
-- SPDX-License-Identifier: Apache-2.0

-- Full NIH research security training content per NOT-OD-26-017.
-- Replaces the stub content added in migration 000004.
-- Completion writes attest:research-security-training=true + expiry tag.

UPDATE training_modules SET
  title = 'NIH Research Security (NOT-OD-26-017)',
  description = 'Required for all key personnel on NIH-funded awards. Covers the 12-month training requirement, foreign affiliation disclosure, and reporting obligations. Must be completed within 12 months of award and renewed annually.',
  estimated_minutes = 45,
  content = '{"sections": [
  {
    "id": "requirements",
    "title": "NIH Research Security Requirements",
    "type": "text",
    "content": "NIH Notice NOT-OD-26-017 requires all **key personnel** on NIH-funded awards to complete research security training within 12 months of award issuance and every 12 months thereafter.\n\n**Who is key personnel?**\nKey personnel are individuals who contribute in a substantive, measurable way to the scientific development or execution of the project, including: Principal Investigators, Co-Investigators, postdoctoral researchers, and staff scientists whose role is integral to the project.\n\nGraduate students, undergraduate students, and consultants are generally NOT key personnel — but check your award terms and consult your Sponsored Programs office if uncertain.\n\n**The 12-month cycle:**\nYour first training must be completed within 12 months of when you become key personnel on an NIH award. Renewal is required every 12 months. Your institution is responsible for tracking compliance; in the Provabl suite, completion is tracked via the attest:research-security-training IAM tag.\n\n**Consequences of non-compliance:**\nNon-compliance is a condition of award that NIH may enforce. Consequences can include: requirement to provide evidence of completion, suspension of funding, and in cases of repeat non-compliance, award termination. NIH treats this as a compliance obligation, not a suggestion.\n\n**What this training covers:**\nThe three areas this module covers — NIH requirements, foreign affiliation disclosure, and reporting obligations — together address the primary research security risks NIH has identified in federally funded research."
  },
  {
    "id": "disclosure",
    "title": "Foreign Affiliation Disclosure",
    "type": "text",
    "content": "One of the most important obligations under research security policy is **full and accurate disclosure of foreign affiliations and support**.\n\n**What must be disclosed:**\n- All foreign institutional affiliations and positions, including honorary and unpaid roles\n- Participation in foreign talent recruitment programs (this includes programs that pay stipends, provide laboratory resources, or offer career development benefits)\n- Foreign grants, contracts, or other research support — including support from foreign governments, foreign companies, or foreign non-profits\n- Intellectual property agreements with foreign entities that affect your research\n\n**When to disclose:**\nDisclosure is required at the time of application AND throughout the award period. If your affiliations or support change after award — you gain a new affiliation, join a talent program, or receive new foreign support — you must notify your institution promptly. Your institution then evaluates whether NIH notification is required.\n\n**What is a foreign talent recruitment program?**\nForeign talent recruitment programs are organized efforts by foreign governments to recruit individuals who work in or with access to US research, often with financial incentives. Chinese government programs (Thousand Talents, Young Thousand Talents, and their successor programs) are the most frequently cited examples, but programs from other countries may also be covered. These programs are not inherently improper — but participation must be disclosed.\n\n**Why this matters:**\nFailure to disclose is treated as research misconduct, not merely an administrative oversight. Cases where researchers have failed to disclose talent program participation or foreign funding have resulted in criminal charges, grant debarment, and institution-level sanctions. The consequences are severe and have affected researchers who did not realize disclosure was required.\n\n**When in doubt, disclose.**\nYour institution''s Sponsored Programs or Research Compliance office can help you determine what must be disclosed. It is always better to over-disclose than to under-disclose."
  },
  {
    "id": "reporting",
    "title": "Reporting Obligations",
    "type": "text",
    "content": "Research security policy creates specific reporting obligations when you encounter activities that may represent inappropriate foreign influence on your research.\n\n**What to report:**\nYou should report to your institution (NOT directly to NIH) if you experience:\n- Requests from foreign entities for unpublished research results, data, or technical information\n- Approaches from representatives of foreign governments or talent programs seeking to recruit you\n- Requests for access to laboratory space, equipment, or personnel that seem unusual or are from unknown foreign parties\n- Pressure to share pre-publication findings with foreign collaborators outside normal scientific channels\n- Unexplained access to your computer systems, research data, or laboratory by unknown parties\n\n**How to report:**\nReport to your institution''s Research Security, Research Compliance, or Export Control office. Do NOT report directly to NIH or federal agencies — your institution manages the initial review and determines if further reporting is required.\n\n**Non-reporting risk vs. false-positive reporting:**\nYou may hesitate to report an incident because you are unsure whether it is actually a problem, or because you do not want to create difficulties for a colleague or collaborator. This hesitation is understandable but misguided. The cost of a false-positive report is low — your institution investigates and determines no action is needed. The cost of failing to report a genuine incident can be catastrophic for you, your colleagues, and your institution. Report, and let the experts determine significance.\n\n**Confidentiality:**\nMost institutions have confidential reporting mechanisms. You can report concerns without identifying yourself. If you are concerned about retaliation, speak with your institution''s ombudsperson.\n\n**After you report:**\nYour institution will conduct an initial assessment. They may ask follow-up questions. They will determine whether federal agency notification (NIH, FBI, or others) is required. You are not responsible for making that determination."
  }
 ],
 "quiz": [
  {
    "id": "q1",
    "question": "Under NOT-OD-26-017, how soon after becoming key personnel on an NIH award must you complete research security training?",
    "options": [
      "30 days",
      "6 months",
      "12 months",
      "Before the next renewal period"
    ],
    "correct": 2,
    "explanation": "NOT-OD-26-017 requires completion within 12 months of when you become key personnel on an NIH award, and renewal every 12 months thereafter."
  },
  {
    "id": "q2",
    "question": "You receive an invitation to join a program offered by a foreign university that will provide an annual stipend and access to laboratory facilities. This program is described as a ''talent introduction program.'' What should you do?",
    "options": [
      "Accept the program if the research it supports does not overlap with your NIH work",
      "Decline all such programs as a matter of policy",
      "Disclose the program to your institution regardless of whether you accept",
      "Disclose only if you accept and the program provides more than $25,000 in value"
    ],
    "correct": 2,
    "explanation": "Foreign talent recruitment programs must be disclosed to your institution regardless of whether you accept. The fact that you received the invitation and are considering it must be disclosed. Your institution will evaluate whether NIH notification is required. There is no dollar threshold for disclosure of these programs."
  },
  {
    "id": "q3",
    "question": "Your foreign collaborator asks you to share your unpublished dataset before peer review. This is a collaborator you have worked with for years at a reputable foreign university. What should you do?",
    "options": [
      "Share the data as you would with any scientific collaborator",
      "Share the data but note in your records that it was shared before publication",
      "Evaluate whether this request is within normal scientific collaboration norms and consult your institution if uncertain",
      "Refuse all pre-publication data sharing with foreign collaborators"
    ],
    "correct": 2,
    "explanation": "Normal scientific collaboration includes pre-publication data sharing with established collaborators. However, if the request feels unusual (e.g., from someone outside your regular collaboration, for data outside your joint project scope, or with unusual urgency), you should consult your Research Security office. The key question is whether the request fits the pattern of legitimate scientific collaboration."
  },
  {
    "id": "q4",
    "question": "You discover that a colleague in your lab has been attending meetings of a foreign government-sponsored program without disclosing this to your institution. Your colleague says the meetings are just networking and there is no formal affiliation. What is your obligation?",
    "options": [
      "No obligation — this is your colleague''s personal matter",
      "Talk to your colleague and encourage them to self-report",
      "Report this to your institution''s Research Security office",
      "Wait to see if the colleague receives formal compensation before reporting"
    ],
    "correct": 2,
    "explanation": "This is a situation where reporting to your institution is appropriate. Your institution''s Research Security office is equipped to assess whether this constitutes a disclosure violation and what steps are needed. The participation in a foreign government program — even informal networking — typically requires disclosure. You are not making an accusation; you are flagging something that needs institutional assessment."
  },
  {
    "id": "q5",
    "question": "In the Provabl qualify system, completing this training writes which IAM tag to your role?",
    "options": [
      "attest:nih-approval=true",
      "attest:research-security-training=true",
      "attest:coc-check-current=true",
      "attest:itar-training=true"
    ],
    "correct": 1,
    "explanation": "Completing the nih-research-security training module writes attest:research-security-training=true (plus an expiry timestamp) to your IAM role. This tag is read by attest''s Cedar PDP for controls that require research security training compliance."
  }
 ],
 "passing_score": 80,
 "prerequisite_modules": ["security-awareness"],
 "certificate_template": "nih-research-security-completion"
}'
WHERE name = 'nih-research-security';
