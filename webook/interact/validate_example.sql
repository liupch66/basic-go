# 测试双写在 webook/interact/client_test.go 的 TestGRPCDoubleWrite
# 分别 "post"
# localhost:8083/migrator/src_only
# localhost:8083/migrator/src_fist
# localhost:8083/migrator/dst_firs
# localhost:8083/migrator/dst_only
# 即可测试


select * from webook.interacts;
select * from webook_interact.interacts;
# 模拟数据 utime 不要乱写 递增或不变 所以每次测试最好清空数据 truncate
# 前面一定要先 "post" localhost:8083/migrator/src_only 或者
# localhost:8083/migrator/src_first 路由设置好 scheduler 的 pattern
# src_only 全量校验
truncate table webook.interacts;
truncate table webook_interact.interacts;
INSERT INTO webook.`interacts`(`biz_id`, `biz`, `read_cnt`, `collect_cnt`, `like_cnt`, `ctime`, `utime`)
VALUES(1,"test",2658,5900,2925,1734961387826,1734961387826),
      (2,"test",5584,8500,8454,1734961387826,1734961387826),
      (3,"test",6372,719,2760,1734961387826,1734961387826),
      (4,"test",9415,2372,526,1734961387826,1734961387826),
      (5,"test",5867,8214,9816,1734961387826,1734961387826),
      (6,"test",5464,7039,1154,1734961387826,1734961387826),
      (7,"test",3158,138,8781,1734961387826,1734961387826),
      (8,"test",6104,3983,1082,1734961387826,1734961387826),
      (9,"test",6442,4568,5170,1734961387826,1734961387826),
      (10,"test",6348,5176,1229,1734961387826,1734961387826);
# src_only 增量校验
INSERT INTO webook.`interacts`(`biz_id`, `biz`, `read_cnt`, `collect_cnt`, `like_cnt`, `ctime`, `utime`)
VALUES(11,"test",26,59,29,1734961387826,1734961387826);

INSERT INTO webook.`interacts`(`biz_id`, `biz`, `read_cnt`, `collect_cnt`, `like_cnt`, `ctime`, `utime`)
VALUES(12,"test",26,59,29,1734961387826,1734961387826);

INSERT INTO webook.`interacts`(`biz_id`, `biz`, `read_cnt`, `collect_cnt`, `like_cnt`, `ctime`, `utime`)
VALUES(13,"test",26,59,29,1734961387826,1734961387826);

# 前面一定要先 "post" localhost:8083/migrator/dst_first 或者
# localhost:8083/migrator/dst_only 路由设置好 scheduler 的 pattern
# dst_first 全量校验
truncate table webook.interacts;
truncate table webook_interact.interacts;
INSERT INTO webook_interact.`interacts`(`biz_id`, `biz`, `read_cnt`, `collect_cnt`, `like_cnt`, `ctime`, `utime`)
VALUES(1,"test",2658,5900,2925,1734961387826,1734961387826),
      (2,"test",5584,8500,8454,1734961387826,1734961387826),
      (3,"test",6372,719,2760,1734961387826,1734961387826),
      (4,"test",9415,2372,526,1734961387826,1734961387826),
      (5,"test",5867,8214,9816,1734961387826,1734961387826),
      (6,"test",5464,7039,1154,1734961387826,1734961387826),
      (7,"test",3158,138,8781,1734961387826,1734961387826),
      (8,"test",6104,3983,1082,1734961387826,1734961387826),
      (9,"test",6442,4568,5170,1734961387826,1734961387826),
      (10,"test",6348,5176,1229,1734961387826,1734961387826);

# dst_first 增量校验
INSERT INTO webook_interact.`interacts`(`biz_id`, `biz`, `read_cnt`, `collect_cnt`, `like_cnt`, `ctime`, `utime`)
VALUES(11,"test",26,59,29,1734961387826,1734961387826);

INSERT INTO webook_interact.`interacts`(`biz_id`, `biz`, `read_cnt`, `collect_cnt`, `like_cnt`, `ctime`, `utime`)
VALUES(12,"test",26,59,29,1734961387826,1734961387826);

INSERT INTO webook_interact.`interacts`(`biz_id`, `biz`, `read_cnt`, `collect_cnt`, `like_cnt`, `ctime`, `utime`)
VALUES(13,"test",26,59,29,1734961387826,1734961387826);