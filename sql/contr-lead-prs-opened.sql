with tot as (
  select
    count(distinct case when a.type = 'authored-commit' then a.sourceId when a.type in ('committed-commit', 'co-authored-commit') then a.sourceParentId else null end) as commits,
    count(distinct split_part(a.url, '#', 1)) filter (where a.type = 'issues-opened') as issues_opened,
    count(distinct split_part(a.url, '#', 1)) filter (where a.type = 'issues-closed') as issues_closed,
    count(a.id) filter (where a.type in ('pull_request-comment','pull_request-review-thread-comment')) as pr_comments,
    count(distinct split_part(a.url, '#', 1)) filter (where a.type = 'pull_request-reviewed') as pr_reviews,
    count(distinct split_part(a.url, '#', 1)) filter (where a.type = 'pull_request-closed') as prs_closed,
    count(distinct split_part(a.url, '#', 1)) filter (where a.type = 'pull_request-merged') as prs_merged,
    count(distinct split_part(a.url, '#', 1)) filter (where a.type = 'pull_request-opened') as prs_opened,
    count(distinct case when a.type = 'authored-commit' then a.sourceId when a.type in ('committed-commit', 'co-authored-commit') then a.sourceParentId else a.id::text end) as contributions,
    count(distinct (memberId, platform, username)) filter (where a.type in ('authored-commit', 'committed-commit', 'co-authored-commit')) as committers,
    count(distinct (memberId, platform, username)) filter (where a.type = 'issues-opened') as issue_openers,
    count(distinct (memberId, platform, username)) filter (where a.type = 'issues-closed') as issue_closers,
    count(distinct (memberId, platform, username)) filter (where a.type in ('pull_request-comment','pull_request-review-thread-comment')) as pr_commenters,
    count(distinct (memberId, platform, username)) filter (where a.type = 'pull_request-reviewed') as pr_reviewers,
    count(distinct (memberId, platform, username)) filter (where a.type = 'pull_request-closed') as pr_closers,
    count(distinct (memberId, platform, username)) filter (where a.type = 'pull_request-merged') as pr_mergers,
    count(distinct (memberId, platform, username)) filter (where a.type = 'pull_request-opened') as pr_openers,
    count(distinct (memberId, platform, username)) as contributors
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
), curr as (
  select
    m.logo_url,
    a.memberId,
    a.platform,
    a.username,
    m.is_bot,
    count(distinct case when a.type = 'authored-commit' then a.sourceId when a.type in ('committed-commit', 'co-authored-commit') then a.sourceParentId else null end) as commits,
    count(distinct split_part(a.url, '#', 1)) filter (where a.type = 'issues-opened') as issues_opened,
    count(distinct split_part(a.url, '#', 1)) filter (where a.type = 'issues-closed') as issues_closed,
    count(a.id) filter (where a.type in ('pull_request-comment','pull_request-review-thread-comment')) as pr_comments,
    count(distinct split_part(a.url, '#', 1)) filter (where a.type = 'pull_request-reviewed') as pr_reviews,
    count(distinct split_part(a.url, '#', 1)) filter (where a.type = 'pull_request-closed') as prs_closed,
    count(distinct split_part(a.url, '#', 1)) filter (where a.type = 'pull_request-merged') as prs_merged,
    count(distinct split_part(a.url, '#', 1)) filter (where a.type = 'pull_request-opened') as prs_opened,
    count(distinct case when a.type = 'authored-commit' then a.sourceId when a.type in ('committed-commit', 'co-authored-commit') then a.sourceParentId else a.id::text end) as contributions
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
  group by
    m.logo_url,
    a.memberId,
    a.platform,
    a.username,
    m.is_bot
), prev as (
  select
    m.logo_url,
    a.memberId,
    a.platform,
    a.username,
    m.is_bot,
    count(distinct case when a.type = 'authored-commit' then a.sourceId when a.type in ('committed-commit', 'co-authored-commit') then a.sourceParentId else null end) as commits,
    count(distinct split_part(a.url, '#', 1)) filter (where a.type = 'issues-opened') as issues_opened,
    count(distinct split_part(a.url, '#', 1)) filter (where a.type = 'issues-closed') as issues_closed,
    count(a.id) filter (where a.type in ('pull_request-comment','pull_request-review-thread-comment')) as pr_comments,
    count(distinct split_part(a.url, '#', 1)) filter (where a.type = 'pull_request-reviewed') as pr_reviews,
    count(distinct split_part(a.url, '#', 1)) filter (where a.type = 'pull_request-closed') as prs_closed,
    count(distinct split_part(a.url, '#', 1)) filter (where a.type = 'pull_request-merged') as prs_merged,
    count(distinct split_part(a.url, '#', 1)) filter (where a.type = 'pull_request-opened') as prs_opened,
    count(distinct case when a.type = 'authored-commit' then a.sourceId when a.type in ('committed-commit', 'co-authored-commit') then a.sourceParentId else a.id::text end) as contributions
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
    and a.timestamp >= {{date_from}}::timestamp - ({{date_to}}::timestamp - {{date_from}}::timestamp)
    and a.timestamp < {{date_to}}::timestamp - ({{date_to}}::timestamp - {{date_from}}::timestamp)
    and m.is_bot {{is_bot}}
    and p.project_slug = '{{project_slug}}'
    and (a.memberId, a.platform, a.username) in (select memberId, platform, username from curr)
  group by
    m.logo_url,
    a.memberId,
    a.platform,
    a.username,
    m.is_bot
)
select
  'prs-opened' as metric,
  c.logo_url,
  c.memberId,
  c.platform,
  c.username,
  c.is_bot,
  c.contributions,
  coalesce(p.contributions, 0) as prev_contributions,
  t.contributions as all_contributions,
  100.0 * (c.contributions::float / t.contributions::float) as percent_total,
  case p.contributions is null when true then 100.0 else 100.0 * ((c.contributions - p.contributions)::float / p.contributions::float) end as change_from_previous,
  t.contributors as all_contributors
from
  tot t,
  curr c
left join
  prev p
on
  c.memberid = p.memberid
  and c.platform = p.platform
  and c.username = p.username
order by
  c.contributions desc
limit
  {{limit}}
