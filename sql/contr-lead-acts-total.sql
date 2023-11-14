select
  count(distinct case when a.type = 'authored-commit' then a.sourceId when a.type in ('committed-commit','co-authored-commit') then a.sourceParentId else a.id::text end) as contributions
from
  activities a
join
  mv_members m
on
  a.memberId = m.id 
join
  mv_subprojects p
on
  a.segmentId = p.id
where
  (
    a.type in (
      'issue-comment', 'issues-closed', 'issues-opened',
      'pull_request-closed', 'pull_request-comment', 'pull_request-merged',
      'pull_request-opened', 'pull_request-review-thread-comment', 'pull_request-reviewed'
    ) or (
      a.type in ('committed-commit', 'co-authored-commit', 'authored-commit')
      and a.attributes->>'isMainBranch' = 'true'
    )
  )
  and a.tenantId = {{tenant_id}}
  and a.deletedAt is null
  and a.timestamp >= {{date_from}}
  and a.timestamp < {{date_to}}
  and m.is_bot {{is_bot}}
  and p.project_slug = '{{project_slug}}'
