with tot as (
  select
    count(distinct split_part(a.url, '#', 1)) as contributions,
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
    a.type = 'pull_request-closed'
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
    count(distinct split_part(a.url, '#', 1)) as contributions
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
    a.type = 'pull_request-closed'
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
    count(distinct split_part(a.url, '#', 1)) as contributions
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
    a.type = 'pull_request-closed'
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
  'prs-closed' as metric,
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
