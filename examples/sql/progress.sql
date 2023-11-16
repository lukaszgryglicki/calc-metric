select
  project_slug,
  time_range,
  metric,
  min(date_from) as date_from,
  max(date_to) as date_to,
  max(row_number) as max_row,
  count(*) as records,
  max(last_calculated_at) as calculated_at,
  max(all_contributions) as all_contributions,
  max(all_contributors) as all_contributors,
  avg(percent_total) as avg_percent_total,
  avg(change_from_previous) as avg_change_from_previous
from
  metric_contr_lead_nbot
group by
  project_slug,
  time_range,
  metric
order by
  calculated_at desc
limit
  60
;
