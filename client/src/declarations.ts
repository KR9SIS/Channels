export type SDPMsg = {
  type: 'video-offer' | 'video-answer'
  name: string
  target: string
  sdp: string
}

export type newICECandidate = {
  type: 'new-ice-candidate'
  target: string
  candidate: string
}
