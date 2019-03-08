package msg

import (
	"fmt"
	"path"
	"strings"

	"github.com/coredns/coredns/plugin/pkg/dnsutil"

	"github.com/miekg/dns"
)

// Path converts a domainname to an etcd path. If s looks like service.staging.skydns.local.,
// the resulting key will be /skydns/local/skydns/staging/service .
func Path(s, prefix string) string {
	l := dns.SplitDomainName(s)
	for i, j := 0, len(l)-1; i < j; i, j = i+1, j-1 {
		l[i], l[j] = l[j], l[i]
	}
	return path.Join(append([]string{"/" + prefix + "/"}, l...)...)
}

// Domain is the opposite of Path.
func Domain(s string) string {
	l := strings.Split(s, "/")
	// start with 1, to strip /skydns
	for i, j := 1, len(l)-1; i < j; i, j = i+1, j-1 {
		l[i], l[j] = l[j], l[i]
	}
	return dnsutil.Join(l[1 : len(l)-1])
}

// PathWithWildcard ascts as Path, but if a name contains wildcards (* or any), the name will be
// chopped of before the (first) wildcard, and we do a highler evel search and
// later find the matching names.  So service.*.skydns.local, will look for all
// services under skydns.local and will later check for names that match
// service.*.skydns.local.  If a wildcard is found the returned bool is true.
func PathWithWildcard(s, prefix string) (string, bool) {
	l := dns.SplitDomainName(s)
	for i, j := 0, len(l)-1; i < j; i, j = i+1, j-1 {
		l[i], l[j] = l[j], l[i]
	}
	for i, k := range l {
		if k == "*" || k == "any" {
			return path.Join(append([]string{"/" + prefix + "/"}, l[:i]...)...), true
		}
	}
	return path.Join(append([]string{"/" + prefix + "/"}, l...)...), false
}

// RootPathSubDomain is used to get the upper level etcd path of the current sub-domain.
// for example:
//    name: sub1.a1.lb.rancher.cloud.
//    sub-domain-path: /skydns/_sub-domain/a1/lb/rancher/cloud/sub1
//    upper-level-path(root): /skydns/_sub-domain/a1/lb/rancher/cloud
func RootPathSubDomain(s []string, wildcardBound int8, prefix string) string {
	s = s[(int8(len(s)) - wildcardBound):]
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	path := strings.Join(s, ".")
	return Path(path, prefix+"/_sub-domain")
}

// PathSubDomain is used to get the real path which stored in etcd.
// for example:
//    name: sub1.a1.lb.rancher.cloud.
//    root(upper-level-path): /skydns/_sub-domain/a1/lb/rancher/cloud/
//    sub-domain-path: /skydns/_sub-domain/a1/lb/rancher/cloud/sub1
func PathSubDomain(root string, wildcardBound int8, s []string) string {
	slice := s[:(int8(len(s)) - wildcardBound)]
	for i, j := 0, len(slice)-1; i < j; i, j = i+1, j-1 {
		slice[i], slice[j] = slice[j], slice[i]
	}
	return fmt.Sprintf("%s/%s", root, strings.Join(slice, "/"))
}
