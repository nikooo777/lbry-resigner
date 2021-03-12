# lbry-resigner

# Requirements
You will need lbrynet version 0.88.0 or greater, credits in at least one account you control, a valid channel you control.

# Usage
```
automatically re-sign streams with invalid signing channels. Only pass --from to dry run

Usage:
  resigner [flags]

Flags:
      --from string              claimID of the old channel
      --funding-account string   id of the funding account used to pay for the transaction
  -h, --help                     help for resigner
      --to string                claimID of the new channel to sign streams with

```

# Example output:

```
./stream_resigner --from bdebe36e74fe497aa27f7f83368dd8aac2d364c1 --to 67f4261d06c541cb7c8e449b734e21bfb79c8b56 --funding-account bRs7qhuK4jV1xUb1VpVM3xkW6ReLcRsHZD
INFO[0000] ------accounts------                         
INFO[0000] 12) account id: bRs7qhuK4jV1xUb1VpVM3xkW6ReLcRsHZD - balance: 10.5 - account name: Account #bRs7qhuK4jV1xUb1VpVM3xkW6ReLcRsHZD 
INFO[0000] 15) account id: bJvkVYwgu1uQsAYioAqxenwdLSZCAKcdxR - balance: 0 - account name: Account #bJvkVYwgu1uQsAYioAqxenwdLSZCAKcdxR 
INFO[0000] 16) account id: bawNLd9NRyyDyz7C7jBDDEy1Y4AUTSPEsm - balance: 3 - account name: Account #bawNLd9NRyyDyz7C7jBDDEy1Y4AUTSPEsm 
INFO[0000] 17) account id: bMVaq2jU9KTwYDY9ihLXBphiF9vn7NCdwc - balance: 0.5 - account name: Account #bMVaq2jU9KTwYDY9ihLXBphiF9vn7NCdwc 
INFO[0001] ------unspent channels------                 
INFO[0001] @Swiss-Experiments - a0f1f8e50dbc4d104edc5640081a71d19cb98702 - 36125a998c6d03acaba501b8e32f0edaedbd980595969ea106b4a205d3af01cc:0 - url:"https://thumbnails.lbry.com/UCNQfQvFMPnInwsU_iGYArJQ"  
INFO[0001] @thumbnails-test - 67f4261d06c541cb7c8e449b734e21bfb79c8b56 - d3c93cb5e0e324d1fc08857fa4d46390916bbaefca3f3a896ef4c17e78b53b38:0 - <nil> 
INFO[0001] ------spent channels------                   
INFO[0001] @Swiss-Experiments - a0f1f8e50dbc4d104edc5640081a71d19cb98702 - 343ae3646a23a44d59cc7abb562b28b73393fca3394135165b5976a66af84c9e:0 - url:"https://thumbnails.lbry.com/UCNQfQvFMPnInwsU_iGYArJQ"  
INFO[0001] @Swiss-Experiments - a0f1f8e50dbc4d104edc5640081a71d19cb98702 - edca63a0e580716a29a475e8e67eff7a6f456893de4d14016c59e5906b8771f0:0 - url:"https://thumbnails.lbry.com/UCNQfQvFMPnInwsU_iGYArJQ"  
INFO[0001] @Swiss-Experiments - a0f1f8e50dbc4d104edc5640081a71d19cb98702 - bed8165554e1497c5e342e096b107672da047f7e551bc6113e2caeb408a3d1f4:0 - url:"https://thumbnails.lbry.com/UCNQfQvFMPnInwsU_iGYArJQ"  
INFO[0001] @Swiss-Experiments - a0f1f8e50dbc4d104edc5640081a71d19cb98702 - 77811a3f51f9c0dbad30b51105276494170b227ca2dc68b651868fd9bcfa938d:0 - url:"https://thumbnails.lbry.com/UCNQfQvFMPnInwsU_iGYArJQ"  
INFO[0001] @Swiss-Experiments - a0f1f8e50dbc4d104edc5640081a71d19cb98702 - dad62c3ba994360da962078ca2a96d54be14ef631ce7d50fdd9f9dbc3b72eaf0:0 - url:"https://thumbnails.lbry.com/UCNQfQvFMPnInwsU_iGYArJQ"  
INFO[0001] @Swiss-Experiments - a0f1f8e50dbc4d104edc5640081a71d19cb98702 - e2c557f57e379160caa86ac9b78b7bbdcaa095b9920ffb6ac8fb88d96dcb35f2:0 - url:"https://thumbnails.lbry.com/UCNQfQvFMPnInwsU_iGYArJQ"  
INFO[0001] @Swiss-Experiments - a0f1f8e50dbc4d104edc5640081a71d19cb98702 - 6b412367a2b1a75a81059fe1fd56b81c7c96bdb90f041e932c95638c5b9e398a:0 - <nil> 
INFO[0001] ------streams without valid signatures------ 
INFO[0001] kawasaki-z650 - invalid channel:  (5b405630077096bf6ff9a3d0079d7abf22bb2a03) 
INFO[0001] swiss-spring-2020 - invalid channel:  (5b405630077096bf6ff9a3d0079d7abf22bb2a03) 
INFO[0001] valentine - invalid channel:  (5b405630077096bf6ff9a3d0079d7abf22bb2a03) 
INFO[0001] goodnight-youtube - invalid channel:  (5b405630077096bf6ff9a3d0079d7abf22bb2a03) 
INFO[0001] devops-colored - invalid channel:  (5b405630077096bf6ff9a3d0079d7abf22bb2a03) 
INFO[0001] rick-and-morty-S04E01 - invalid channel: @RickAndMorty (3e684554030ddb58481f0612997e921962077164) 
INFO[0001] ytprova2-NV4VDOjJb7o - invalid channel:  (bdebe36e74fe497aa27f7f83368dd8aac2d364c1) 
INFO[0001] ytprova2-Adko0l6z42c - invalid channel:  (bdebe36e74fe497aa27f7f83368dd8aac2d364c1) 
INFO[0001] ytprova2-PsLKmlbqwYM - invalid channel:  (bdebe36e74fe497aa27f7f83368dd8aac2d364c1) 
INFO[0002] successful update. TXID: 3f6ab06cce0ff227363cfaa1691962ccab50cc56c11eb1d7cad03b14dd1c2b4d 
INFO[0003] successful update. TXID: a8f10d2806f01f54ff86fcee3ed0f7467c98ea29bbbbcde503d866178ba26bc6 
INFO[0003] successful update. TXID: 3f553f295cd078f1d09b69a03c0720d85a2188ed3af8102923c7865fc72e8dcb
```

# todo

- [ ] Allow resigning valid streams