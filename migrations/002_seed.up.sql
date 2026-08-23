INSERT INTO users(id,display_name,login,password_hash,roles,unit_id,created_at) VALUES
('rep-demo','演示代表','representative',decode('$2a$10$0gYF7B8WUBs52ysSOy3m2u1E9D3pLhLAwLp8mBGYP9gPc0MsmCf1q','escape'),ARRAY['representative'],'',now()),
('unit-demo','承办员','unit_officer',decode('$2a$10$0gYF7B8WUBs52ysSOy3m2u1E9D3pLhLAwLp8mBGYP9gPc0MsmCf1q','escape'),ARRAY['unit_officer'],'unit-education',now()),
('supervisor-demo','督办员','supervisor',decode('$2a$10$0gYF7B8WUBs52ysSOy3m2u1E9D3pLhLAwLp8mBGYP9gPc0MsmCf1q','escape'),ARRAY['supervisor'],'',now())
ON CONFLICT DO NOTHING;
INSERT INTO assignment_rules(id,category,unit_id,priority,active) VALUES
('rule-education','education','unit-education',100,true),
('rule-transport','transport','unit-transport',100,true),
('rule-community','community','unit-community',100,true)
ON CONFLICT DO NOTHING;
