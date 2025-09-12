# multi transaction (distributed tracing)

## 중요사항은 아래와 같습니다. 

### trace.Start, trace.StartWithContext 를 사용할 경우, 
    *  request의 Header를 분석하는 로직이 필요합니다. 
        trace.UpdateMtrace()
    * 외부 api 연결시 whatap 에서 추가할 Header를 별도로 추가하는 로직이 필요합니다. 
        trace.GetMtrace() 
        header.Set()
### trace.StartWithContext
    trace.StartWithContext(ctx context.Context, name)
    이 함수는 조금 특이합니다. (사용자가 혼동할 가능성이 있는 API입니다. )
    전달하는 ctx 는 미리 whatap api 를 통해서 생성된 context(내부에 whatap 이름으로 whatap의 tracecontext가 존재)를 전달해야 합니다. 

    ctx, _ = trace.NewTraceContext(ctx)
	ctx, _ = trace.StartWithContext(ctx, fmt.Sprintf("%s_%d", "/trace3", depth))

### distributed_tracing/server/server.go 
    
    3.1 Makefile
      - bin/mtrace 를 생성합니다. 
    3.1 bin/run.sh
      - whatap.conf
        라이센스 및 서버 호스트 정보를 수정합니다. 
      - 실행
        서버포트를 8080, 8081, 8082, 8083 을 4개의 Server를 실행합니다. 

      - Test url 
        *  http://localhost:8080/trace1_0
            -> 멀티트랜잭션 연계
                trace1_0 -> trace1_1 -> trace1_2 -> trac1_3

        * http://localhost:8080/trace2_0
             -> 멀티트랜잭션 연계
                trace2_0 -> trace2_1 -> trace2_2 -> trac2_3
        * http://localhost:8080/trace3_0
            -> 멀티트랜잭션 연계
                trace3_0 -> trace3_1 -> trace3_2 -> trac3_3

  3.1 trace1
    * trace.StartWithRequest 관련 함수 예제

  3.2 trace2
    * trace.Start 관련 함수 예제

  3.3 trace3 
    * trace.StartWithContext 관련 함수 예제



