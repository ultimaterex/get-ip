# FireHOL blocklist-ipsets — HTTP feed URLs (`BLOCKLIST_URLS`)

This inventory lists **154** feed files in the [firehol/blocklist-ipsets](https://github.com/firehol/blocklist-ipsets) repo `master` branch (snapshot used for this doc). Files change upstream — confirm names on GitHub before relying on them.

**Pattern:**

```text
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/<filename>
```

**Official Spamhaus text feeds** (not in the FireHOL tree — often used for DROP):

- `https://www.spamhaus.org/drop/drop.txt`
- `https://www.spamhaus.org/drop/edrop.txt`

**Caution:** Many feeds overlap; some are ISP/game/ad oriented or extremely large. See [blocklist-examples.md](./blocklist-examples.md). get-ip caps each URL download at **64 MiB**.

---

## Full URL list (one per line)

```text
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/abuseipdb_1d.ipset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/abuseipdb_30d.ipset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/bds_atif.ipset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/bitcoin_nodes.ipset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/bitcoin_nodes_1d.ipset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/bitcoin_nodes_30d.ipset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/bitcoin_nodes_7d.ipset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/blocklist_de.ipset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/blocklist_de_apache.ipset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/blocklist_de_bots.ipset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/blocklist_de_bruteforce.ipset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/blocklist_de_ftp.ipset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/blocklist_de_imap.ipset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/blocklist_de_mail.ipset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/blocklist_de_sip.ipset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/blocklist_de_ssh.ipset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/blocklist_de_strongips.ipset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/blocklist_net_ua.ipset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/botscout.ipset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/botscout_1d.ipset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/botscout_30d.ipset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/botscout_7d.ipset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/botvrij_dst.ipset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/botvrij_src.ipset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/bruteforceblocker.ipset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/c2_tracker.ipset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/ciarmy.ipset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/cidr_report_bogons.netset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/cleantalk.ipset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/cleantalk_1d.ipset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/cleantalk_30d.ipset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/cleantalk_7d.ipset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/cleantalk_new.ipset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/cleantalk_new_1d.ipset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/cleantalk_new_30d.ipset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/cleantalk_new_7d.ipset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/cleantalk_top20.ipset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/cleantalk_updated.ipset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/cleantalk_updated_1d.ipset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/cleantalk_updated_30d.ipset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/cleantalk_updated_7d.ipset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/cta_cryptowall.ipset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/cybercrime.ipset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/cybercure.ipset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/darklist_de.netset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/dm_tor.ipset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/dshield.netset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/dshield_1d.netset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/dshield_30d.netset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/dshield_7d.netset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/et_block.netset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/et_compromised.ipset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/et_dshield.netset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/et_spamhaus.netset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/et_tor.ipset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/feodo.ipset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/feodo_badips.ipset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/firehol_abusers_1d.netset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/firehol_abusers_30d.netset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/firehol_anonymous.netset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/firehol_level1.netset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/firehol_level2.netset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/firehol_level3.netset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/firehol_level4.netset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/firehol_proxies.netset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/firehol_webclient.netset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/firehol_webserver.netset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/gpf_comics.ipset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/graphiclineweb.netset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/greensnow.ipset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/iblocklist_abuse_palevo.netset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/iblocklist_abuse_spyeye.netset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/iblocklist_abuse_zeus.netset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/iblocklist_ciarmy_malicious.netset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/iblocklist_cidr_report_bogons.netset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/iblocklist_cruzit_web_attacks.netset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/iblocklist_isp_aol.netset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/iblocklist_isp_att.netset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/iblocklist_isp_cablevision.netset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/iblocklist_isp_charter.netset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/iblocklist_isp_comcast.netset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/iblocklist_isp_embarq.netset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/iblocklist_isp_qwest.netset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/iblocklist_isp_sprint.netset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/iblocklist_isp_suddenlink.netset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/iblocklist_isp_twc.netset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/iblocklist_isp_verizon.netset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/iblocklist_malc0de.netset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/iblocklist_onion_router.netset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/iblocklist_org_activision.netset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/iblocklist_org_apple.netset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/iblocklist_org_blizzard.netset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/iblocklist_org_crowd_control.netset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/iblocklist_org_electronic_arts.netset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/iblocklist_org_joost.netset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/iblocklist_org_linden_lab.netset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/iblocklist_org_logmein.netset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/iblocklist_org_ncsoft.netset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/iblocklist_org_nintendo.netset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/iblocklist_org_pandora.netset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/iblocklist_org_pirate_bay.netset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/iblocklist_org_punkbuster.netset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/iblocklist_org_riot_games.netset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/iblocklist_org_sony_online.netset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/iblocklist_org_square_enix.netset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/iblocklist_org_steam.netset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/iblocklist_org_ubisoft.netset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/iblocklist_org_xfire.netset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/iblocklist_pedophiles.netset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/iblocklist_spamhaus_drop.netset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/iblocklist_yoyo_adservers.netset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/maltrail_scanners.ipset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/myip.ipset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/php_commenters.ipset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/php_commenters_1d.ipset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/php_commenters_30d.ipset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/php_commenters_7d.ipset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/php_dictionary.ipset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/php_dictionary_1d.ipset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/php_dictionary_30d.ipset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/php_dictionary_7d.ipset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/php_harvesters.ipset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/php_harvesters_1d.ipset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/php_harvesters_30d.ipset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/php_harvesters_7d.ipset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/php_spammers.ipset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/php_spammers_1d.ipset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/php_spammers_30d.ipset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/php_spammers_7d.ipset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/sblam.ipset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/socks_proxy.ipset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/socks_proxy_1d.ipset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/socks_proxy_30d.ipset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/socks_proxy_7d.ipset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/spamhaus_drop.netset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/spamhaus_edrop.netset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/sslproxies.ipset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/sslproxies_1d.ipset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/sslproxies_30d.ipset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/sslproxies_7d.ipset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/stopforumspam.ipset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/stopforumspam_180d.ipset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/stopforumspam_1d.ipset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/stopforumspam_30d.ipset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/stopforumspam_365d.ipset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/stopforumspam_7d.ipset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/stopforumspam_90d.ipset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/stopforumspam_toxic.netset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/tor_exits.ipset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/tor_exits_1d.ipset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/tor_exits_30d.ipset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/tor_exits_7d.ipset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/vxvault.ipset
https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/yoyo_adservers.ipset
```

---

## Regenerating

1. Refresh `documentation/firehol-filenames.txt` from the [GitHub API directory listing](https://api.github.com/repos/firehol/blocklist-ipsets/contents?ref=master) (keep only `*.netset` / `*.ipset` names).
2. Run: `node documentation/gen-firehol-md.mjs`
